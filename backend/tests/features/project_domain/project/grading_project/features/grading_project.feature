Feature: Grading-Project API
  As a user
  I want to setup grading project for a project
  So that I can put, read and download grading project for a project if I have permission

  Background:
    Given I have users:
      | id | name               | username | password | roles            |
      | 11 | AdminUser          | user01   | 123456   | admin            |
      | 12 | APIReader          | user02   | 123456   | api_reader       |
      | 13 | Contestant         | user03   | 123456   | contestant       |
      | 14 | AdminAndContestant | user04   | 123456   | admin,contestant |
    And I have images:
      | id | tag               | mount_path    | description  | is_ready | createdAt           | updatedAt           |
      | 1  | nginx:latest      | /var/www/html | description1 | 0        | 2021-09-01T00:00:00 | 2021-09-01T00:00:00 |
      | 2  | mysql:8.0.26      | /var/www/html | description2 | 1        | 2021-09-01T00:00:00 | 2021-09-01T00:00:00 |
      | 3  | test-nginx:latest | /var/www/html | description3 | 0        | 2021-09-01T00:00:00 | 2021-09-01T00:00:00 |
      | 4  | test-mysql:8.0.26 | /var/www/html | description4 | 1        | 2021-09-01T00:00:00 | 2021-09-01T00:00:00 |
      | 5  | test-nginx:test-2 | /var/www/html | description5 | 1        | 2021-09-01T00:00:00 | 2021-09-01T00:00:00 |
    And I have projects:
      | id | name           | description               | imageTag       | env | testImageTag      | testEnv | resultParser | createdAt           | updatedAt           |
      | 11 | Test-Project-1 | Test-Project-1 Description| nginx:latest   | a=b | test-nginx:latest |         |  | 2021-09-01 00:00:00 | 2021-09-01 00:00:00 |
      | 12 | Test-Project-2 | Test-Project-2 Description| mysql:8.0.26   | c=d | test-mysql:8.0.26 |         |  | 2021-09-01 00:00:00 | 2021-09-01 00:00:00 |
      | 13 | Test-Project-3 | Test-Project-3 Description| nginx:latest   |     | test-nginx:test-2 |         |  | 2021-09-01 00:00:00 | 2021-09-01 00:00:00 |
  Scenario: Successful handle grading-project with "AdminUser" user
    Given I am logged in as "user01"
    When I send a "PUT" request to "/api/v1/projects/11/grading-project" with the following formdata:
      | key              | value             |
      | project          | {nginx project}   |
    Then the response code should be 200
    And the response should match json:
      """
      {
        "success": true
      }
      """
    And I can download the grading-project with "GET" request to "/api/v1/projects/11/grading-project/download"
    When I send a "GET" request to "/api/v1/projects/11/grading-project/files":
    Then the response code should be 200
    And the response should match json:
      """
      {
          "success": true,
          "data": [
              "Dockerfile",
              "conf.d/default.conf"
          ]
      }
      """
    When I send a "GET" request to "/api/v1/projects/11/grading-project/files/RG9ja2VyZmlsZQ==":
    Then the response code should be 200
    And the response should match json:
      """
      {
          "success": true,
          "data": "RlJPTSBuZ2lueA0KDQpDT1BZIC4vY29uZi5kIC9ldGMvbmdpbngvY29uZi5kDQo="
      }
      """
    When I send a "GET" request to "/api/v1/projects/11/grading-project/files/Y29uZi5kL2RlZmF1bHQuY29uZg==":
    Then the response code should be 200
    And the response should match json:
      """
      {
          "success": true,
          "data": "c2VydmVyIHsNCiAgICBsaXN0ZW4gICAgICAgODA7DQogICAgbGlzdGVuICBbOjpdOjgwOw0KICAgIHNlcnZlcl9uYW1lICBsb2NhbGhvc3Q7DQoNCiAgICAjYWNjZXNzX2xvZyAgL3Zhci9sb2cvbmdpbngvaG9zdC5hY2Nlc3MubG9nICBtYWluOw0KDQogICAgbG9jYXRpb24gLyB7DQogICAgICAgIHJvb3QgICAvdmFyL3d3dzsNCiAgICAgICAgaW5kZXggIGluZGV4Lmh0bWwgaW5kZXguaHRtOw0KICAgIH0NCg0KICAgICNlcnJvcl9wYWdlICA0MDQgICAgICAgICAgICAgIC80MDQuaHRtbDsNCg0KICAgICMgcmVkaXJlY3Qgc2VydmVyIGVycm9yIHBhZ2VzIHRvIHRoZSBzdGF0aWMgcGFnZSAvNTB4Lmh0bWwNCiAgICAjDQogICAgZXJyb3JfcGFnZSAgIDUwMCA1MDIgNTAzIDUwNCAgLzUweC5odG1sOw0KICAgIGxvY2F0aW9uID0gLzUweC5odG1sIHsNCiAgICAgICAgcm9vdCAgIC91c3Ivc2hhcmUvbmdpbngvaHRtbDsNCiAgICB9DQoNCiAgICAjIHByb3h5IHRoZSBQSFAgc2NyaXB0cyB0byBBcGFjaGUgbGlzdGVuaW5nIG9uIDEyNy4wLjAuMTo4MA0KICAgICMNCiAgICAjbG9jYXRpb24gfiBcLnBocCQgew0KICAgICMgICAgcHJveHlfcGFzcyAgIGh0dHA6Ly8xMjcuMC4wLjE7DQogICAgI30NCg0KICAgICMgcGFzcyB0aGUgUEhQIHNjcmlwdHMgdG8gRmFzdENHSSBzZXJ2ZXIgbGlzdGVuaW5nIG9uIDEyNy4wLjAuMTo5MDAwDQogICAgIw0KICAgICNsb2NhdGlvbiB+IFwucGhwJCB7DQogICAgIyAgICByb290ICAgICAgICAgICBodG1sOw0KICAgICMgICAgZmFzdGNnaV9wYXNzICAgMTI3LjAuMC4xOjkwMDA7DQogICAgIyAgICBmYXN0Y2dpX2luZGV4ICBpbmRleC5waHA7DQogICAgIyAgICBmYXN0Y2dpX3BhcmFtICBTQ1JJUFRfRklMRU5BTUUgIC9zY3JpcHRzJGZhc3RjZ2lfc2NyaXB0X25hbWU7DQogICAgIyAgICBpbmNsdWRlICAgICAgICBmYXN0Y2dpX3BhcmFtczsNCiAgICAjfQ0KDQogICAgIyBkZW55IGFjY2VzcyB0byAuaHRhY2Nlc3MgZmlsZXMsIGlmIEFwYWNoZSdzIGRvY3VtZW50IHJvb3QNCiAgICAjIGNvbmN1cnMgd2l0aCBuZ2lueCdzIG9uZQ0KICAgICMNCiAgICAjbG9jYXRpb24gfiAvXC5odCB7DQogICAgIyAgICBkZW55ICBhbGw7DQogICAgI30NCn0NCg=="
      }
      """

  
  # Everything an uploaded project can contain that used to break extraction,
  # in one archive: dotfiles the old rule silently deleted, a folder named
  # after the wrapper folder that made unwrap fail with "file exists", and the
  # __MACOSX/.DS_Store noise that must still be dropped. The last part is not
  # cosmetic — extracting __MACOSX would leave two entries at the archive root
  # and silently switch unwrap off, burying the Dockerfile one level too deep.
  # Project 13, not 11: grading projects persist on disk across scenarios in
  # this feature, and the read/download scenarios below assert on the nginx
  # project uploaded to 11.
  Scenario: Successful put grading-project containing dotfiles and a folder named after the project
    Given I am logged in as "user01"
    When I send a "PUT" request to "/api/v1/projects/13/grading-project" with the following formdata:
      | key              | value             |
      | project          | {tricky project}  |
    Then the response code should be 200
    And the response should match json:
      """
      {
        "success": true
      }
      """
    When I send a "GET" request to "/api/v1/projects/13/grading-project/files":
    Then the response code should be 200
    And the response should match json:
      """
      {
          "success": true,
          "data": [
              ".env",
              ".htaccess",
              "Dockerfile",
              "myproject/index.html"
          ]
      }
      """
    # The dotfile is not just listed, its contents actually survived.
    When I send a "GET" request to "/api/v1/projects/13/grading-project/files/LmVudg==":
    Then the response code should be 200
    And the response should match json:
      """
      {
          "success": true,
          "data": "U0VDUkVUPWNvbnRlc3QK"
      }
      """

  Scenario: Successful download grading-project with "APIReader" user
    Given I am logged in as "user02"
    And I can download the grading-project with "GET" request to "/api/v1/projects/11/grading-project/download"

  Scenario: Successful read grading-project file with "APIReader" user
    Given I am logged in as "user02"
    When I send a "GET" request to "/api/v1/projects/11/grading-project/files":
    Then the response code should be 200
    And the response should match json:
      """
      {
          "success": true,
          "data": [
              "Dockerfile",
              "conf.d/default.conf"
          ]
      }
      """
    When I send a "GET" request to "/api/v1/projects/11/grading-project/files/RG9ja2VyZmlsZQ==":
    Then the response code should be 200
    And the response should match json:
      """
      {
          "success": true,
          "data": "RlJPTSBuZ2lueA0KDQpDT1BZIC4vY29uZi5kIC9ldGMvbmdpbngvY29uZi5kDQo="
      }
      """
    When I send a "GET" request to "/api/v1/projects/11/grading-project/files/Y29uZi5kL2RlZmF1bHQuY29uZg==":
    Then the response code should be 200
    And the response should match json:
      """
      {
          "success": true,
          "data": "c2VydmVyIHsNCiAgICBsaXN0ZW4gICAgICAgODA7DQogICAgbGlzdGVuICBbOjpdOjgwOw0KICAgIHNlcnZlcl9uYW1lICBsb2NhbGhvc3Q7DQoNCiAgICAjYWNjZXNzX2xvZyAgL3Zhci9sb2cvbmdpbngvaG9zdC5hY2Nlc3MubG9nICBtYWluOw0KDQogICAgbG9jYXRpb24gLyB7DQogICAgICAgIHJvb3QgICAvdmFyL3d3dzsNCiAgICAgICAgaW5kZXggIGluZGV4Lmh0bWwgaW5kZXguaHRtOw0KICAgIH0NCg0KICAgICNlcnJvcl9wYWdlICA0MDQgICAgICAgICAgICAgIC80MDQuaHRtbDsNCg0KICAgICMgcmVkaXJlY3Qgc2VydmVyIGVycm9yIHBhZ2VzIHRvIHRoZSBzdGF0aWMgcGFnZSAvNTB4Lmh0bWwNCiAgICAjDQogICAgZXJyb3JfcGFnZSAgIDUwMCA1MDIgNTAzIDUwNCAgLzUweC5odG1sOw0KICAgIGxvY2F0aW9uID0gLzUweC5odG1sIHsNCiAgICAgICAgcm9vdCAgIC91c3Ivc2hhcmUvbmdpbngvaHRtbDsNCiAgICB9DQoNCiAgICAjIHByb3h5IHRoZSBQSFAgc2NyaXB0cyB0byBBcGFjaGUgbGlzdGVuaW5nIG9uIDEyNy4wLjAuMTo4MA0KICAgICMNCiAgICAjbG9jYXRpb24gfiBcLnBocCQgew0KICAgICMgICAgcHJveHlfcGFzcyAgIGh0dHA6Ly8xMjcuMC4wLjE7DQogICAgI30NCg0KICAgICMgcGFzcyB0aGUgUEhQIHNjcmlwdHMgdG8gRmFzdENHSSBzZXJ2ZXIgbGlzdGVuaW5nIG9uIDEyNy4wLjAuMTo5MDAwDQogICAgIw0KICAgICNsb2NhdGlvbiB+IFwucGhwJCB7DQogICAgIyAgICByb290ICAgICAgICAgICBodG1sOw0KICAgICMgICAgZmFzdGNnaV9wYXNzICAgMTI3LjAuMC4xOjkwMDA7DQogICAgIyAgICBmYXN0Y2dpX2luZGV4ICBpbmRleC5waHA7DQogICAgIyAgICBmYXN0Y2dpX3BhcmFtICBTQ1JJUFRfRklMRU5BTUUgIC9zY3JpcHRzJGZhc3RjZ2lfc2NyaXB0X25hbWU7DQogICAgIyAgICBpbmNsdWRlICAgICAgICBmYXN0Y2dpX3BhcmFtczsNCiAgICAjfQ0KDQogICAgIyBkZW55IGFjY2VzcyB0byAuaHRhY2Nlc3MgZmlsZXMsIGlmIEFwYWNoZSdzIGRvY3VtZW50IHJvb3QNCiAgICAjIGNvbmN1cnMgd2l0aCBuZ2lueCdzIG9uZQ0KICAgICMNCiAgICAjbG9jYXRpb24gfiAvXC5odCB7DQogICAgIyAgICBkZW55ICBhbGw7DQogICAgI30NCn0NCg=="
      }
      """

  # A project that exists but never had grading files uploaded is an ordinary
  # state, and the download says so with a 404 — not a 500 with an error-level
  # log entry for a file that was never supposed to be there.
  Scenario: Download grading-project of a project without uploaded files is not found
    Given I have no grading project files
    And I am logged in as "user01"
    When I send a "GET" request to "/api/v1/projects/12/grading-project/download":
    Then the response code should be 404

  Scenario: Failed put grading-project with project not found
    Given I am logged in as "user01"
    When I send a "PUT" request to "/api/v1/projects/111/grading-project" with the following formdata:
      | key              | value             |
      | project          | {nginx project}   |
    Then the response code should be 404
    And the response should match json:
      """
      {
        "success": false,
        "message": "project not found",
        "details": null
      }
      """

  Scenario: Unauthorized put grading-project
    When I send a "PUT" request to "/api/v1/projects/11/grading-project" with the following formdata:
      | key              | value             |
      | project          | {nginx project}   |
    Then the response code should be 401
    And the response should match json:
      """
      {
        "success": false,
        "message": "Unauthorized",
        "details": null
      }
      """

  Scenario: Unauthorized read grading-project file
    When I send a "GET" request to "/api/v1/projects/11/grading-project/files":
    Then the response code should be 401
    And the response should match json:
      """
      {
        "success": false,
        "message": "Unauthorized",
        "details": null
      }
      """
    When I send a "GET" request to "/api/v1/projects/11/grading-project/files/RG9ja2VyZmlsZQ==":
    Then the response code should be 401
    And the response should match json:
      """
      {
        "success": false,
        "message": "Unauthorized",
        "details": null
      }
      """
    
  Scenario: Forbidden put grading-project with "APIReader" user
    Given I am logged in as "user02"
    When I send a "PUT" request to "/api/v1/projects/11/grading-project" with the following formdata:
      | key              | value             |
      | project          | {nginx project}   |
    Then the response code should be 403
    And the response should match json:
      """
      {
        "success": false,
        "message": "Forbidden",
        "details": null
      }
      """

  Scenario: Forbidden read grading-project file with "Contestant" user
    Given I am logged in as "user03"
    When I send a "GET" request to "/api/v1/projects/11/grading-project/files":
    Then the response code should be 403
    And the response should match json:
      """
      {
        "success": false,
        "message": "Forbidden",
        "details": null
      }
      """
    When I send a "GET" request to "/api/v1/projects/11/grading-project/files/RG9ja2VyZmlsZQ==":
    Then the response code should be 403
    And the response should match json:
      """
      {
        "success": false,
        "message": "Forbidden",
        "details": null
      }
      """

  Scenario: Forbidden download grading-project with "Contestant" user
    Given I am logged in as "user03"
    And I send a "GET" request to "/api/v1/projects/11/grading-project/download":
    Then the response code should be 403
    And the response should match json:
      """
      {
        "success": false,
        "message": "Forbidden",
        "details": null
      }
      """
