import "@typespec/http";
import "@typespec/openapi";

using TypeSpec.OpenAPI;
using TypeSpec.Http;
using Shared;

@tag("Csrf")
@route("/")
namespace Common.CsrfSpec;

model CsrfToken {
    token: string;
    ttl: int32;
}
alias CsrfTokenResponse = APIResponse<CsrfToken> | APIInternalServerErrorResponse;

interface Interface {
    @get
    @route("/csrf-token")
    @operationId("getCsrfToken")
    getCsrfToken(): CsrfTokenResponse;

    @delete
    @route("/csrf-token")
    @operationId("deleteCsrfToken")
    deleteCsrfToken(): APINoDataResponse;
}
