import "./model.tsp";
import "@typespec/http";
import "@typespec/openapi";

using TypeSpec.OpenAPI;
using TypeSpec.Http;
using Shared;

@tag("Users")
@route("/users")
namespace Admin.UserSpec;

model CreateUserRequest {
    name: string;
    username: string;
    password: string;
    roles?: int32[];
}

model UpdateUserRequest {
    name?: string;
    password?: string;
    roles?: int32[];
}

model ImportUsersRequest {
    users: bytes;
}

alias ImportUsersResponse = APICreateResponse<int32>;

interface Interface
    extends CqrsResourcesInterface<
            "User",
            User,
            User,
            CreateUserRequest,
            UpdateUserRequest
        > {
    @post
    @usePermission(AuthAdmin)
    @route("/import")
    @operationId("importUsers")
    importUsers(
        @header contentType: "multipart/form-data",
        @body payload: ImportUsersRequest,
    ): ImportUsersResponse;
}
