import "./model.tsp";
import "@typespec/http";
import "@typespec/openapi";

using TypeSpec.OpenAPI;
using TypeSpec.Http;
using Shared;

@tag("Auth")
@route("/auth")
namespace Common.AuthSpec;

model LoginRequest {
    username: string;
    password: string;
}

/**
 * 429。登入的連續失敗保護(由中介層直接回應):與其他 429 不同,沒有
 * details 也沒有 retryAfterSeconds — 稍候片刻再試即可。401 與 429 是
 * 刻意區分的:前者是帳密錯誤,後者是嘗試過於頻繁。
 */
@error
model LoginRateLimitedResponse {
    @statusCode statusCode: 429;
    message: string;
    success: boolean = false;
}

alias LoginResponse = APIResponse<Token> | APIBadRequestResponse| APIUnauthorizedResponse | LoginRateLimitedResponse | APIInternalServerErrorResponse;
alias LogoutResponse = APINoDataResponse | APIBadRequestResponse | APIUnauthorizedResponse | APIInternalServerErrorResponse;
alias getProfileResponse = APIResponse<Profile> | APIBadRequestResponse | APIUnauthorizedResponse | APIInternalServerErrorResponse;
alias refreshTokenResponse = APIResponse<Token> | APIBadRequestResponse | APIUnauthorizedResponse | APIInternalServerErrorResponse;

interface Interface {
    @post
    @useAuth(CsrfToken)
    @route("/login")
    @operationId("login")
    login(@body payload: LoginRequest): LoginResponse;

    @post
    @useAuth(RefreshToken)
    @route("/logout")
    @operationId("logout")
    logout(): LogoutResponse;

    @get
    @usePermission(Authorization)
    @route("/profile")
    @operationId("getProfile")
    getProfile(): getProfileResponse;

    @post
    @useAuth(RefreshToken)
    @route("/refresh-token")
    @operationId("refreshToken")
    refreshToken(): refreshTokenResponse;
}
