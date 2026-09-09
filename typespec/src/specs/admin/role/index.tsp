import "./model.tsp";
import "@typespec/http";
import "@typespec/openapi";

using TypeSpec.OpenAPI;
using TypeSpec.Http;
using Shared;

@tag("Roles")
@route("/roles")
namespace Admin.RoleSpec;

interface Interface {
    @get
    @usePermission(AuthRead)
    @route("/")
    @operationId("listRole")
    listRole(): APIListResponse<Role>;
}
