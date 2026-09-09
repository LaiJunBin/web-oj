Feature: Composite Bundle Export / Preview / Import API
  As a contest organiser moving a whole sitting to a new venue
  I want one file that carries the contest, its problems and their images
  So that setting up the venue is one upload instead of a sequence I must order correctly

  # The composite bundle defines no fourth format: projects/<slug>/ inside it
  # IS a project bundle, images/<slug>/ IS an image bundle. This suite
  # therefore leans on the part suites for field-level fidelity and pins what
  # is new here — the manifest, the preview's judgement, the import order and
  # the report.
  #
  # Import is two steps by design. Preview shows what would be created, what
  # already exists and what forcing would affect, and writes nothing; import
  # refuses whatever preview would have flagged. The two cannot drift because
  # import runs the preview as its gate.

  Background:
    Given I have no grading project files
    And I have users:
      | id | name       | username | password | roles      |
      | 11 | AdminUser  | user01   | 123456   | admin      |
      | 12 | APIReader  | user02   | 123456   | api_reader |
      | 13 | Contestant | user03   | 123456   | contestant |
    And I have images:
      | id | tag           | mount_path    | description | is_ready | createdAt           | updatedAt           |
      | 1  | comp-img:v1   | /var/www/html | 環境        | 1        | 2021-09-01T00:00:00 | 2021-09-01T00:00:00 |
    And setup "nginx.zip" project into "comp-img:v1" image
    And I have projects:
      | id | name       | description | imageTag    | env | testImageTag | testEnv | resultParser | createdAt           | updatedAt           |
      | 11 | Comp-Alpha | A           | comp-img:v1 |     | comp-img:v1  |         |              | 2021-09-01 00:00:00 | 2021-09-01 00:00:00 |
    And I have grading-items:
      | id | projectId | name | description | score | order | scoringStrategy | outputFormat | resultParser | label | createdAt           | updatedAt           |
      | 41 | 11        | g1   |             | 10    | 1     | increase        | text         |              | 1     | 2021-09-01T00:00:00 | 2021-09-01T00:00:00 |
    And I have testcases:
      | id | gradingItemId | name  | command | score | isSample | createdAt           | updatedAt           |
      | 51 | 41            | g1-t1 | true    | 10    | 0        | 2021-09-01T00:00:00 | 2021-09-01T00:00:00 |
    And I have contests:
      | id | name      | description | isPublic | activateTime        | deactivateTime      | createdAt           | updatedAt           |
      | 21 | Comp-2026 | 完整匯出檔來源 | 0        | 2026-09-01 00:00:00 | 2026-09-15 00:00:00 | 2021-09-01 00:00:00 | 2021-09-01 00:00:00 |
    And I have contest project associations:
      | id | contestId | projectId | name  | isPublic | judgePriorityStrategy | activateTime        | deactivateTime      | startTime           | endTime             |
      | 31 | 21        | 11        | 第一題 | 1        | default               | 2026-09-01 00:00:00 | 2026-09-15 00:00:00 | 2026-09-01 09:00:00 | 2026-09-01 17:00:00 |

  # ─────────────────────────── Export ───────────────────────────

  Scenario: Exporting a contest as a composite bundle carries everything it needs
    Given I am logged in as "user01"
    When I export the bundle with a "GET" request to "/api/v1/bundles/export/contests/21?mode=source"
    Then the exported bundle should contain:
      | path                            |
      | manifest.json                   |
      | contest.json                    |
      | projects/p11/project.json       |
      | projects/p11/meta.json          |
      | images/comp-img-v1/image.json   |
      | images/comp-img-v1/context.zip  |
    And the exported bundle file "manifest.json" should match json:
      """
      {
        "formatVersion": 1,
        "kind": "composite",
        "images": ["comp-img-v1"],
        "projects": ["p11"],
        "hasContest": true,
        "exportedAt": "{exportedAt}"
      }
      """

  Scenario: An archive-mode export of an unbuilt image is refused with the build list
    Given I have images:
      | id | tag             | mount_path    | description | is_ready | createdAt           | updatedAt           |
      | 2  | comp-unbuilt:v1 | /var/www/html | 沒建置       | 0        | 2021-09-01T00:00:00 | 2021-09-01T00:00:00 |
    And I have projects:
      | id | name      | description | imageTag        | env | testImageTag    | testEnv | resultParser | createdAt           | updatedAt           |
      | 12 | Comp-Beta | B           | comp-unbuilt:v1 |     | comp-unbuilt:v1 |         |              | 2021-09-01 00:00:00 | 2021-09-01 00:00:00 |
    And I have contest project associations:
      | id | contestId | projectId | name  | isPublic | judgePriorityStrategy | activateTime        | deactivateTime      | startTime           | endTime             |
      | 32 | 21        | 12        | 第二題 | 1        | default               | 2026-09-01 00:00:00 | 2026-09-15 00:00:00 | 2026-09-01 09:00:00 | 2026-09-01 17:00:00 |
    And I am logged in as "user01"
    When I send a "GET" request to "/api/v1/bundles/export/contests/21?mode=archive":
    Then the response code should be 400

  Scenario: Contestants may not export
    Given I am logged in as "user03"
    When I send a "GET" request to "/api/v1/bundles/export/contests/21?mode=source":
    Then the response code should be 403

  # ─────────────────────────── Preview ───────────────────────────

  # The venue already holds an identical comp-img:v1 (same build context, so
  # the fingerprints agree) — the preview says so and clears the import.
  Scenario: Previewing a bundle against an installation that already has its image
    Given I am logged in as "user01"
    When I export the bundle with a "GET" request to "/api/v1/bundles/export/contests/21?mode=source"
    And I import the exported bundle with a "POST" request to "/api/v1/bundles/preview"
    Then the response code should be 200
    And the response should match json:
      """
      {
        "success": true,
        "data": {
          "formatVersion": 1,
          "images": [
            { "tag": "comp-img:v1", "mode": "source", "status": "identical", "usedByProjects": [] }
          ],
          "projects": [
            { "name": "Comp-Alpha", "status": "new", "missingImageTags": [] }
          ],
          "contest": { "name": "Comp-2026", "projectSlots": 1, "unresolvedProjects": [] },
          "blockers": []
        }
      }
      """

  Scenario: Previewing writes nothing
    Given I am logged in as "user01"
    When I export the bundle with a "GET" request to "/api/v1/bundles/export/contests/21?mode=source"
    And I import the exported bundle with a "POST" request to "/api/v1/bundles/preview"
    Then the response code should be 200
    And there should be 1 projects
    And there should be 1 contests

  Scenario: A file that is not a composite bundle is rejected at preview
    Given I am logged in as "user01"
    When I import a bundle with a "POST" request to "/api/v1/bundles/preview" whose "not-manifest.json" is:
      """
      {}
      """
    Then the response code should be 400
    And the response should match json:
      """
      {
        "success": false,
        "message": "{message}",
        "details": ["匯出檔裡沒有 manifest.json，這看起來不是「完整匯出」產生的檔案"]
      }
      """

  # ─────────────────────────── Import ───────────────────────────

  # The whole point of the feature: one upload recreates the sitting. The
  # image is identical so it is reused; the project and contest are created.
  Scenario: Importing a composite bundle recreates the sitting in one step
    Given I am logged in as "user01"
    When I export the bundle with a "GET" request to "/api/v1/bundles/export/contests/21?mode=source"
    And I import the exported bundle with a "POST" request to "/api/v1/bundles/import"
    Then the response code should be 200
    And the response should match json:
      """
      {
        "success": true,
        "data": {
          "createdImages": [],
          "reusedImages": ["comp-img:v1"],
          "queuedBuilds": [],
          "createdProjectIds": [1e+308],
          "contestId": 1e+308,
          "unresolvedProjects": [],
          "warnings": []
        }
      }
      """
    And there should be 2 projects
    And there should be 2 contests

  Scenario: Contestants may not import
    Given I am logged in as "user03"
    When I import a bundle with a "POST" request to "/api/v1/bundles/import" whose "manifest.json" is:
      """
      {}
      """
    Then the response code should be 403
    And there should be 1 projects

  Scenario: Read-only users may not import
    Given I am logged in as "user02"
    When I import a bundle with a "POST" request to "/api/v1/bundles/import" whose "manifest.json" is:
      """
      {}
      """
    Then the response code should be 403
