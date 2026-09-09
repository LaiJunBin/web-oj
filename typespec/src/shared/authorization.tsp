import "@typespec/http";
import "../lib/usePermission.js";

using TypeSpec.Http;

namespace Shared;

model CsrfToken is ApiKeyAuth<ApiKeyLocation.header, "csrf-token">;
model Authorization is ApiKeyAuth<ApiKeyLocation.cookie, "access_token">;
model RefreshToken is ApiKeyAuth<ApiKeyLocation.cookie, "refresh_token">;
model UserID is ApiKeyAuth<ApiKeyLocation.header, "UserID">;
model PermissionRead is ApiKeyAuth<ApiKeyLocation.header, "PermissionRead">;
model PermissionAdmin is ApiKeyAuth<ApiKeyLocation.header, "PermissionAdmin">;
model PermissionContestant is ApiKeyAuth<ApiKeyLocation.header, "PermissionContestant">;


alias WithUserID = [Authorization,UserID];
alias AuthRead = WithUserID | PermissionRead;
alias AuthAdmin = WithUserID | PermissionAdmin;
alias AuthContestant = WithUserID | PermissionContestant;

