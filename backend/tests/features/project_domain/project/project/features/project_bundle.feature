Feature: Project Bundle Export / Import API
  As a contest organiser setting up a new venue
  I want to move a problem between installations as one file
  So that nobody has to retype its grading items and testcases, or edit the seeder

  # A problem is spread over eight tables and two directories on disk. The
  # bundle is the only artefact that carries all of it, so these scenarios
  # assert on the WHOLE of project.json rather than the field under discussion:
  # a field the exporter forgets is invisible to a partial assertion, and shows
  # up much later as a contest that scores differently at the new venue.
  #
  # The round trip is the centrepiece. Export and import checked separately can
  # both look right while disagreeing about one field; checked against each
  # other they cannot.

  Background:
    # The storage directory is not reset between scenarios, and an export reads
    # whatever is on disk for the project it is given. Without this, a feature
    # that ran earlier in the same suite decides what these bundles contain.
    Given I have no grading project files
    And I have users:
      | id | name               | username | password | roles            |
      | 11 | AdminUser          | user01   | 123456   | admin            |
      | 12 | APIReader          | user02   | 123456   | api_reader       |
      | 13 | Contestant         | user03   | 123456   | contestant       |
    And I have images:
      | id | tag               | mount_path    | description  | is_ready | createdAt           | updatedAt           |
      | 1  | nginx:latest      | /var/www/html | description1 | 1        | 2021-09-01T00:00:00 | 2021-09-01T00:00:00 |
      | 2  | test-nginx:latest | /var/www/html | description3 | 1        | 2021-09-01T00:00:00 | 2021-09-01T00:00:00 |
    And I have projects:
      | id | name           | description                | imageTag     | env      | testImageTag      | testEnv  | resultParser | createdAt           | updatedAt           |
      | 11 | Test-Project-1 | Test-Project-1 Description | nginx:latest | FOO=bar  | test-nginx:latest | BAZ=qux  | auto         | 2021-09-01 00:00:00 | 2021-09-01 00:00:00 |
      | 12 | Empty-Project  | No children at all         | nginx:latest |          | test-nginx:latest |          |              | 2021-09-01 00:00:00 | 2021-09-01 00:00:00 |
    And I have dynamic fields:
      | id | projectId | type    | name    | label    | defaultValue | isRequired | createdAt           | updatedAt           |
      | 21 | 11        | text    | repoUrl | Repo 網址 | https://     | 1          | 2021-09-01T00:00:00 | 2021-09-01T00:00:00 |
      | 22 | 11        | boolean | isDraft | 草稿      |              | 0          | 2021-09-01T00:00:00 | 2021-09-01T00:00:00 |
    And I have lifecycles:
      | id | projectId | name | description | type       | runsOn  | command | createdAt           | updatedAt           |
      | 31 | 11        | 安裝 | 安裝相依     | before_all | project | npm ci  | 2021-09-01T00:00:00 | 2021-09-01T00:00:00 |
    And I have grading-items:
      | id | projectId | name | description | score | order | scoringStrategy | outputFormat | resultParser | label | createdAt           | updatedAt           |
      | 41 | 11        | g1   | g1-desc     | 10    | 1     | decrease        | text         |              | 1     | 2021-09-01T00:00:00 | 2021-09-01T00:00:00 |
      | 42 | 11        | g2   |             | 5     | 2     | increase        | html         | stdout       | 2     | 2021-09-01T00:00:00 | 2021-09-01T00:00:00 |
    And I have testcases:
      | id | gradingItemId | name  | command    | score | isSample | createdAt           | updatedAt           |
      | 51 | 41            | g1-t1 | echo t1    | 4     | 1        | 2021-09-01T00:00:00 | 2021-09-01T00:00:00 |
      | 52 | 41            | g1-t2 | echo t2    | 6     | 0        | 2021-09-01T00:00:00 | 2021-09-01T00:00:00 |
      | 53 | 42            | g2-t1 | echo g2-t1 | 5     | 0        | 2021-09-01T00:00:00 | 2021-09-01T00:00:00 |
    # A dependency pointing at a testcase under ANOTHER grading item, which is
    # the case a per-item key scheme would get wrong.
    And I have testcase depends:
      | testcaseId | dependId |
      | 53         | 51       |
    And I have grading-item depends:
      | gradingItemId | testcaseId |
      | 42            | 52         |
    And I have assets:
      | id                                   | projectId | name     | description | size | isEnabled | createdAt           | updatedAt           |
      | 11111111-1111-1111-1111-111111111111 | 11        | spec.zip | 規格書       | 1094 | 1         | 2021-09-01T00:00:00 | 2021-09-01T00:00:00 |

  # ─────────────────────────── Export ───────────────────────────

  Scenario: Exporting a project writes every part of it into the bundle
    Given I am logged in as "user01"
    When I export the project bundle with a "GET" request to "/api/v1/projects/11/export"
    Then the exported bundle should contain:
      | path                                 |
      | project.json                         |
      | meta.json                            |
      | assets/1.zip                         |
    And the exported bundle file "project.json" should match json:
      """
      {
        "formatVersion": 1,
        "kind": "project",
        "name": "Test-Project-1",
        "description": "Test-Project-1 Description",
        "imageTag": "nginx:latest",
        "testImageTag": "test-nginx:latest",
        "env": "FOO=bar",
        "testEnv": "BAZ=qux",
        "resultParser": "auto",
        "dynamicFields": [
          { "type": "text", "name": "repoUrl", "label": "Repo 網址", "defaultValue": "https://", "isRequired": true },
          { "type": "boolean", "name": "isDraft", "label": "草稿", "defaultValue": "", "isRequired": false }
        ],
        "lifecycles": [
          { "name": "安裝", "description": "安裝相依", "type": "before_all", "runsOn": "project", "command": "npm ci" }
        ],
        "gradingItems": [
          {
            "label": "1",
            "name": "g1",
            "description": "g1-desc",
            "score": 10,
            "scoringStrategy": "decrease",
            "outputFormat": "text",
            "resultParser": "",
            "dependsOnTestcases": [],
            "testcases": [
              { "key": "g1t1", "name": "g1-t1", "command": "echo t1", "score": 4, "isSample": true, "dependsOnTestcases": [] },
              { "key": "g1t2", "name": "g1-t2", "command": "echo t2", "score": 6, "isSample": false, "dependsOnTestcases": [] }
            ]
          },
          {
            "label": "2",
            "name": "g2",
            "description": "",
            "score": 5,
            "scoringStrategy": "increase",
            "outputFormat": "html",
            "resultParser": "stdout",
            "dependsOnTestcases": ["g1t2"],
            "testcases": [
              { "key": "g2t1", "name": "g2-t1", "command": "echo g2-t1", "score": 5, "isSample": false, "dependsOnTestcases": ["g1t1"] }
            ]
          }
        ],
        "assets": [
          { "name": "spec.zip", "isEnabled": true, "file": "assets/1.zip" }
        ],
        "hasGradingFiles": false
      }
      """

  # Provenance is kept out of project.json so two exports of an unchanged
  # project produce identical bytes and a diff shows what somebody changed.
  Scenario: Provenance lives outside the definition
    Given I am logged in as "user01"
    When I export the project bundle with a "GET" request to "/api/v1/projects/11/export"
    Then the exported bundle file "meta.json" should match json:
      """
      {
        "exportedAt": "{exportedAt}",
        "sourceProjectId": 11
      }
      """

  Scenario: A project with no children still exports
    Given I am logged in as "user01"
    When I export the project bundle with a "GET" request to "/api/v1/projects/12/export"
    Then the exported bundle should contain:
      | path         |
      | project.json |
      | meta.json    |
    And the exported bundle file "project.json" should match json:
      """
      {
        "formatVersion": 1,
        "kind": "project",
        "name": "Empty-Project",
        "description": "No children at all",
        "imageTag": "nginx:latest",
        "testImageTag": "test-nginx:latest",
        "env": "",
        "testEnv": "",
        "resultParser": "",
        "dynamicFields": [],
        "lifecycles": [],
        "gradingItems": [],
        "assets": [],
        "hasGradingFiles": false
      }
      """

  # The grading project is part of the problem, not an attachment: it is copied
  # into the test container at judge time. Once uploaded it rides along in the
  # bundle and the definition says so — the "no children" scenario above pins
  # the other half, where nothing was uploaded and the entry is simply absent.
  Scenario: A project with uploaded grading files exports them in the bundle
    Given I am logged in as "user01"
    When I send a "PUT" request to "/api/v1/projects/12/grading-project" with the following formdata:
      | key     | value           |
      | project | {nginx project} |
    Then the response code should be 200
    When I export the project bundle with a "GET" request to "/api/v1/projects/12/export"
    Then the exported bundle should contain:
      | path              |
      | project.json      |
      | meta.json         |
      | grading-files.zip |
    And the exported bundle file "project.json" should match json:
      """
      {
        "formatVersion": 1,
        "kind": "project",
        "name": "Empty-Project",
        "description": "No children at all",
        "imageTag": "nginx:latest",
        "testImageTag": "test-nginx:latest",
        "env": "",
        "testEnv": "",
        "resultParser": "",
        "dynamicFields": [],
        "lifecycles": [],
        "gradingItems": [],
        "assets": [],
        "hasGradingFiles": true
      }
      """

  Scenario: Read-only users may export
    Given I am logged in as "user02"
    When I export the project bundle with a "GET" request to "/api/v1/projects/11/export"
    Then the exported bundle should contain:
      | path         |
      | project.json |
      | meta.json    |
      | assets/1.zip |

  Scenario: Contestants may not export
    Given I am logged in as "user03"
    When I send a "GET" request to "/api/v1/projects/11/export":
    Then the response code should be 403

  Scenario: Exporting requires a session
    When I send a "GET" request to "/api/v1/projects/11/export":
    Then the response code should be 401

  Scenario: Exporting a project that does not exist
    Given I am logged in as "user01"
    When I send a "GET" request to "/api/v1/projects/999/export":
    Then the response code should be 404

  # ─────────────────────────── Round trip ───────────────────────────

  # The assertion that catches a dropped field. Export and import are checked
  # against each other, so a field either one forgets cannot survive.
  Scenario: A round trip recreates the project
    Given I am logged in as "user01"
    When I export the project bundle with a "GET" request to "/api/v1/projects/11/export"
    And I import the exported bundle with a "POST" request to "/api/v1/projects/import"
    Then the response code should be 200
    And the response should match json:
      """
      {
        "success": true,
        "data": {
          "projectId": 1e+308,
          "name": "{name}",
          "imageTag": "nginx:latest",
          "testImageTag": "test-nginx:latest",
          "missingImageTags": []
        }
      }
      """
    # Exporting the copy must produce the same definition as exporting the
    # original, apart from two deliberate differences. The name is suffixed
    # because one already exists. And env gains a trailing newline: the project
    # entity canonicalises it that way, so an API-created project always has
    # one — the Background seeds the column directly, which is a value the
    # product itself would never store. Everything else that differs is a
    # field one side mishandled.
    When I export the project bundle with a "GET" request to "/api/v1/projects/13/export"
    Then the exported bundle file "project.json" should match json:
      """
      {
        "formatVersion": 1,
        "kind": "project",
        "name": "{name}",
        "description": "Test-Project-1 Description",
        "imageTag": "nginx:latest",
        "testImageTag": "test-nginx:latest",
        "env": "FOO=bar\n",
        "testEnv": "BAZ=qux\n",
        "resultParser": "auto",
        "dynamicFields": [
          { "type": "text", "name": "repoUrl", "label": "Repo 網址", "defaultValue": "https://", "isRequired": true },
          { "type": "boolean", "name": "isDraft", "label": "草稿", "defaultValue": "", "isRequired": false }
        ],
        "lifecycles": [
          { "name": "安裝", "description": "安裝相依", "type": "before_all", "runsOn": "project", "command": "npm ci" }
        ],
        "gradingItems": [
          {
            "label": "1",
            "name": "g1",
            "description": "g1-desc",
            "score": 10,
            "scoringStrategy": "decrease",
            "outputFormat": "text",
            "resultParser": "",
            "dependsOnTestcases": [],
            "testcases": [
              { "key": "g1t1", "name": "g1-t1", "command": "echo t1", "score": 4, "isSample": true, "dependsOnTestcases": [] },
              { "key": "g1t2", "name": "g1-t2", "command": "echo t2", "score": 6, "isSample": false, "dependsOnTestcases": [] }
            ]
          },
          {
            "label": "2",
            "name": "g2",
            "description": "",
            "score": 5,
            "scoringStrategy": "increase",
            "outputFormat": "html",
            "resultParser": "stdout",
            "dependsOnTestcases": ["g1t2"],
            "testcases": [
              { "key": "g2t1", "name": "g2-t1", "command": "echo g2-t1", "score": 5, "isSample": false, "dependsOnTestcases": ["g1t1"] }
            ]
          }
        ],
        "assets": [
          { "name": "spec.zip", "isEnabled": true, "file": "assets/1.zip" }
        ],
        "hasGradingFiles": false
      }
      """

  # A project names the image it needs but never carries it, and creating a
  # project against an image this installation does not have is refused — for
  # an import exactly as it would be for a person filling in the same form.
  # Saying so before anything is written beats failing half way, and leaves the
  # operator with an instruction rather than a project that cannot judge.
  Scenario: Importing is refused when the installation lacks the image
    Given I am logged in as "user01"
    When I import a bundle with a "POST" request to "/api/v1/projects/import" whose project.json is:
      """
      {
        "formatVersion": 1,
        "kind": "project",
        "name": "Hand written",
        "description": "",
        "imageTag": "somewhere-else:latest",
        "testImageTag": "test-nginx:latest",
        "env": "",
        "testEnv": "",
        "resultParser": "",
        "dynamicFields": [],
        "lifecycles": [],
        "gradingItems": [],
        "assets": [],
        "hasGradingFiles": false
      }
      """
    Then the response code should be 400
    And the response should match json:
      """
      {
        "success": false,
        "message": "{message}",
        "details": [
          "映像 \"somewhere-else:latest\" 不存在於本系統，請先在「評分環境管理」建立並建置後再匯入"
        ]
      }
      """
    And there should be 2 projects

  # ─────────────────────────── Import failures ───────────────────────────

  # A bundle may be hand-written, so one upload has to report every fault it
  # can see. Reporting them one at a time turns fixing a file into one round
  # trip per typo.
  Scenario: An invalid bundle is rejected with every fault at once
    Given I am logged in as "user01"
    When I import a bundle with a "POST" request to "/api/v1/projects/import" whose project.json is:
      """
      {
        "formatVersion": 1,
        "kind": "project",
        "name": "",
        "description": "",
        "imageTag": "nginx:latest",
        "testImageTag": "test-nginx:latest",
        "env": "",
        "testEnv": "",
        "resultParser": "",
        "dynamicFields": [],
        "lifecycles": [],
        "gradingItems": [
          {
            "label": "1",
            "name": "g1",
            "description": "",
            "score": 10,
            "scoringStrategy": "average",
            "outputFormat": "text",
            "resultParser": "",
            "dependsOnTestcases": ["nothing-declares-this"],
            "testcases": []
          }
        ],
        "assets": [],
        "hasGradingFiles": false
      }
      """
    Then the response code should be 400
    And the response should match json:
      """
      {
        "success": false,
        "message": "{message}",
        "details": [
          "name: 不可為空",
          "gradingItems[0].scoringStrategy: \"average\" 不是合法的計分方式",
          "gradingItems[0].dependsOnTestcases[0]: 找不到 key 為 \"nothing-declares-this\" 的測資"
        ]
      }
      """
    And there should be 2 projects

  Scenario: A bundle from a newer format version is refused rather than guessed at
    Given I am logged in as "user01"
    When I import a bundle with a "POST" request to "/api/v1/projects/import" whose project.json is:
      """
      {
        "formatVersion": 99,
        "kind": "project",
        "name": "From the future",
        "description": "",
        "imageTag": "nginx:latest",
        "testImageTag": "test-nginx:latest",
        "env": "",
        "testEnv": "",
        "resultParser": "",
        "dynamicFields": [],
        "lifecycles": [],
        "gradingItems": [],
        "assets": [],
        "hasGradingFiles": false
      }
      """
    Then the response code should be 400
    And there should be 2 projects

  # The typos that hurt are on optional fields: isSampel leaves isSample false,
  # the import succeeds, and the mistake surfaces during the contest as a
  # testcase mysteriously hidden from contestants.
  Scenario: A misspelled field is rejected rather than ignored
    Given I am logged in as "user01"
    When I import a bundle with a "POST" request to "/api/v1/projects/import" whose project.json is:
      """
      {
        "formatVersion": 1,
        "kind": "project",
        "name": "Typo",
        "description": "",
        "imageTag": "nginx:latest",
        "testImageTag": "test-nginx:latest",
        "env": "",
        "testEnv": "",
        "resultParser": "",
        "dynamicFields": [],
        "lifecycles": [],
        "gradingItems": [
          {
            "label": "1",
            "name": "g1",
            "description": "",
            "score": 10,
            "scoringStrategy": "increase",
            "outputFormat": "text",
            "resultParser": "",
            "dependsOnTestcases": [],
            "testcases": [
              { "key": "a", "name": "t1", "command": "true", "score": 1, "isSampel": true, "dependsOnTestcases": [] }
            ]
          }
        ],
        "assets": [],
        "hasGradingFiles": false
      }
      """
    Then the response code should be 400
    And there should be 2 projects

  Scenario: An upload that is not a zip is rejected
    Given I am logged in as "user01"
    When I import a bundle with a "POST" request to "/api/v1/projects/import" whose content is:
      """
      this is not a zip file
      """
    Then the response code should be 400
    And there should be 2 projects

  Scenario: Read-only users may not import
    Given I am logged in as "user02"
    When I import a bundle with a "POST" request to "/api/v1/projects/import" whose project.json is:
      """
      {
        "formatVersion": 1,
        "kind": "project",
        "name": "Nope",
        "description": "",
        "imageTag": "nginx:latest",
        "testImageTag": "test-nginx:latest",
        "env": "",
        "testEnv": "",
        "resultParser": "",
        "dynamicFields": [],
        "lifecycles": [],
        "gradingItems": [],
        "assets": [],
        "hasGradingFiles": false
      }
      """
    Then the response code should be 403
    And there should be 2 projects
