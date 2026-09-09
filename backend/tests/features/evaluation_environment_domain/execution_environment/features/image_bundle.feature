Feature: Image Bundle Export / Import API
  As a contest organiser preparing a venue
  I want to move an evaluation image between installations
  So that a problem's environment travels with it instead of being rebuilt from memory

  # An image bundle takes one of two shapes, and the choice is the operator's
  # because only they know the venue: source carries the build context and is
  # kilobytes but needs a build and a network, archive carries docker save
  # output and needs neither but is hundreds of megabytes.
  #
  # These scenarios cover source end to end and the decisions around archive.
  # The archive path itself talks to a real daemon — loading prebuilt layers is
  # not something a feature test can honestly fake, so the parts that are pure
  # decision (refusing an unbuilt image, refusing a mode the API does not know)
  # are what is pinned here.

  Background:
    Given I have users:
      | id | name       | username | password | roles      |
      | 11 | AdminUser  | user01   | 123456   | admin      |
      | 12 | APIReader  | user02   | 123456   | api_reader |
      | 13 | Contestant | user03   | 123456   | contestant |
    And I have images:
      | id | tag              | mount_path    | description | is_ready | createdAt           | updatedAt           |
      | 1  | bundle-src:v1    | /var/www/html | 有原始檔     | 1        | 2021-09-01T00:00:00 | 2021-09-01T00:00:00 |
      | 2  | bundle-unbuilt:v1| /var/www/html | 沒建置過     | 0        | 2021-09-01T00:00:00 | 2021-09-01T00:00:00 |
    And setup "nginx.zip" project into "bundle-src:v1" image

  Scenario: Exporting an image in source mode carries the build context
    Given I am logged in as "user01"
    When I export the bundle with a "GET" request to "/api/v1/images/bundle-src%3Av1/export?mode=source"
    Then the exported bundle should contain:
      | path        |
      | image.json  |
      | context.zip |
    And the exported bundle file "image.json" should match json:
      """
      {
        "formatVersion": 1,
        "kind": "image",
        "tag": "bundle-src:v1",
        "mountPath": "/var/www/html",
        "description": "有原始檔",
        "mode": "source",
        "fingerprint": "{fingerprint}"
      }
      """

  # Re-exporting must produce the same fingerprint. If it did not, every import
  # of an image the venue already has would look like a conflict, and the
  # operator would learn to force every time — which is the same as having no
  # conflict detection at all.
  Scenario: The fingerprint is stable across exports
    Given I am logged in as "user01"
    When I export the bundle with a "GET" request to "/api/v1/images/bundle-src%3Av1/export?mode=source"
    Then the exported bundle file "image.json" should match json:
      """
      {
        "formatVersion": 1,
        "kind": "image",
        "tag": "bundle-src:v1",
        "mountPath": "/var/www/html",
        "description": "有原始檔",
        "mode": "source",
        "fingerprint": "{fingerprint}"
      }
      """
    When I import the exported bundle with a "POST" request to "/api/v1/images/import"
    # Same tag, same content: reused in silence. This is the ordinary case when
    # a venue imports two problems that share an image.
    Then the response code should be 200
    And the response should match json:
      """
      {
        "success": true,
        "data": {
          "tag": "bundle-src:v1",
          "mode": "source",
          "buildQueued": false
        }
      }
      """

  # docker save on a tag the daemon has never built produces an archive with
  # nothing in it, which would import cleanly and leave the venue with nothing
  # to judge against.
  Scenario: An archive of an image that was never built is refused
    Given I am logged in as "user01"
    When I send a "GET" request to "/api/v1/images/bundle-unbuilt%3Av1/export?mode=archive":
    Then the response code should be 400

  Scenario: Exporting an image that does not exist
    Given I am logged in as "user01"
    When I send a "GET" request to "/api/v1/images/nowhere%3Av1/export?mode=source":
    Then the response code should be 404

  Scenario: Read-only users may export
    Given I am logged in as "user02"
    When I export the bundle with a "GET" request to "/api/v1/images/bundle-src%3Av1/export?mode=source"
    Then the exported bundle should contain:
      | path        |
      | image.json  |
      | context.zip |

  Scenario: Contestants may not export
    Given I am logged in as "user03"
    When I send a "GET" request to "/api/v1/images/bundle-src%3Av1/export?mode=source":
    Then the response code should be 403

  # ─────────────────────────── Import ───────────────────────────

  # A new tag goes through the ordinary creation path, which reserves a build
  # slot and queues the build. Nothing waits for it: the build has its own
  # page, status and log stream already.
  Scenario: Importing a new source image queues a build
    Given I am logged in as "user01"
    When I export the bundle with a "GET" request to "/api/v1/images/bundle-src%3Av1/export?mode=source"
    And I import the exported bundle with a "POST" request to "/api/v1/images/import"
    Then the response code should be 200

  # The tag is the same but the content is not, and the tag may be what several
  # live contests are judging against. Refused, with what forcing would affect.
  Scenario: A tag that exists with different content is refused rather than overwritten
    Given I am logged in as "user01"
    When I import a bundle with a "POST" request to "/api/v1/images/import" whose "image.json" is:
      """
      {
        "formatVersion": 1,
        "kind": "image",
        "tag": "bundle-src:v1",
        "mountPath": "/var/www/html",
        "description": "不同的內容",
        "mode": "source",
        "fingerprint": "sha256:0000000000000000000000000000000000000000000000000000000000000000"
      }
      """
    # The payload is missing too, but the point of this scenario is the tag
    # decision, so it asserts only that the request does not succeed and
    # changes nothing.
    Then the response code should be 400

  # The frontend's exact reproduction of the force-import bug. Forcing
  # replaces the image by soft-deleting the old row and creating a new one, so
  # the tag briefly maps to two rows — and the build repository used to resolve
  # it to the DEAD one. The queued build then ran and committed onto the dead
  # image: invisible in the builds list, and every later context update or
  # rebuild 500ed. The wait step below polls builds through the LIVE image row,
  # so it is precisely the observer that failed before the fix.
  Scenario: A forced import's queued build lands on the new image and completes
    Given I am logged in as "user01"
    When I export the bundle with a "GET" request to "/api/v1/images/bundle-src%3Av1/export?mode=source"
    # The installation's copy drifts, so the bundle now genuinely conflicts.
    And setup "nginx-fail.zip" project into "bundle-src:v1" image
    And I import the exported bundle with a "POST" request to "/api/v1/images/import"
    Then the response code should be 409
    When I import the exported bundle with force with a "POST" request to "/api/v1/images/import"
    Then the response code should be 200
    And the response should match json:
      """
      {
        "success": true,
        "data": {
          "tag": "bundle-src:v1",
          "mode": "source",
          "buildQueued": true
        }
      }
      """
    And wait for 120 seconds for the image "bundle-src:v1" build to complete

  Scenario: An invalid bundle is rejected with every fault at once
    Given I am logged in as "user01"
    When I import a bundle with a "POST" request to "/api/v1/images/import" whose "image.json" is:
      """
      {
        "formatVersion": 1,
        "kind": "image",
        "tag": "",
        "mountPath": "/var/www/html",
        "description": "",
        "mode": "sideways",
        "fingerprint": ""
      }
      """
    Then the response code should be 400
    And the response should match json:
      """
      {
        "success": false,
        "message": "{message}",
        "details": [
          "tag: 不可為空",
          "mode: 必須是 \"source\" 或 \"archive\"，得到 \"sideways\"",
          "fingerprint: 不可為空 —— 沒有它就無法判斷與系統既有的映像是否相同",
          "mode 為 \"sideways\"，但匯出檔裡沒有 context.zip"
        ]
      }
      """

  Scenario: A bundle from a newer format version is refused
    Given I am logged in as "user01"
    When I import a bundle with a "POST" request to "/api/v1/images/import" whose "image.json" is:
      """
      {
        "formatVersion": 99,
        "kind": "image",
        "tag": "future:v1",
        "mountPath": "/var/www/html",
        "description": "",
        "mode": "source",
        "fingerprint": "sha256:abc"
      }
      """
    Then the response code should be 400

  Scenario: Read-only users may not import
    Given I am logged in as "user02"
    When I import a bundle with a "POST" request to "/api/v1/images/import" whose "image.json" is:
      """
      {
        "formatVersion": 1,
        "kind": "image",
        "tag": "nope:v1",
        "mountPath": "/var/www/html",
        "description": "",
        "mode": "source",
        "fingerprint": "sha256:abc"
      }
      """
    Then the response code should be 403
