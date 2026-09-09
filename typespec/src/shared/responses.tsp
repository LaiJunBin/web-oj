import "@typespec/http";

using TypeSpec.Http;

namespace Shared;


model APINoDataResponse<StatusCode = 200> {
    @statusCode statusCode: StatusCode;
    success: boolean = true;
}

model APIResponse<T, StatusCode = 200>  {
    @statusCode statusCode: StatusCode;
    data: T;
    success: boolean = true;
}

@error
model APIErrorResponse<StatusCode = 400, Message = "Bad Request"> {
    @statusCode statusCode: StatusCode;
    message: Message;
    details: string[];
    success: boolean = false;
}

model PageLinks {
    first: string;
    last: string;
    next: string;
    previous: string;
}

model Pagination {
    page: int32;
    pageSize: int32;
    totalPages: int32;
    totalItems: int32;
    hasNext: boolean;
    hasPrevious: boolean;
    links: PageLinks;
}

model APIPagedResponse<T> {
    data: {
        items: T;
        pagination: Pagination;
    };
    success: boolean = true;
}


model APIBadRequestResponse is APIErrorResponse<400, "Bad Request">;
model APIUnauthorizedResponse is APIErrorResponse<401, "Unauthorized">;
model APIForbiddenResponse is APIErrorResponse<403, "Forbidden">;
model APINotFoundResponse is APIErrorResponse<404, "Not Found">;
model APIInternalServerErrorResponse
    is APIErrorResponse<500, "Internal Server Error">;
model APINotImplementedResponse is APIErrorResponse<501, "Not Implemented">;

/**
 * 429。與其他錯誤回應多一個 retryAfterSeconds，讓前端能顯示倒數而不必解析訊息字串。
 * 刻意不放進下面的 alias：只有真的有頻率限制的 operation 才該宣告它。
 */
@error
model APITooManyRequestsResponse {
    @statusCode statusCode: 429;
    message: string;
    details: string[];
    success: boolean = false;

    /** 還要等幾秒才能再次操作 */
    retryAfterSeconds: int32;
}

/**
 * 409。上傳的匯出檔與系統既有的東西同名，但內容不同。
 *
 * 與 400 分開是因為前端要據此做不同的事：400 是「這個檔案有問題」，
 * 409 是「檔案沒問題，但會覆蓋既有的東西」——後者該給使用者一個明確的
 * 強制覆蓋選項，而不是要他去解析錯誤訊息判斷。
 *
 * details 逐條說明衝突，並列出受影響的範圍（例如哪些題目正在使用這個映像）。
 * 刻意不放進下面的 alias：只有匯入類 operation 會遇到。
 */
@error
model APIConflictResponse {
    @statusCode statusCode: 409;

    message: string;
    details: string[];
    success: boolean = false;
}

alias APIBaseResponseOrAvailableErrors<T> =
    | T
    | APIBadRequestResponse
    | APIUnauthorizedResponse
    | APIForbiddenResponse
    | APINotFoundResponse
    | APIInternalServerErrorResponse
    | APINotImplementedResponse;

alias APIBaseCreateResponse<T> =
    | T
    | APIBadRequestResponse
    | APIUnauthorizedResponse
    | APIForbiddenResponse
    | APINotFoundResponse
    | APIInternalServerErrorResponse
    | APINotImplementedResponse;

alias APIBaseGetResponse<T> =
    | T
    | APIBadRequestResponse
    | APIUnauthorizedResponse
    | APIForbiddenResponse
    | APINotFoundResponse
    | APIInternalServerErrorResponse
    | APINotImplementedResponse;

alias APIBaseListResponse<T> =
    | T
    | APIBadRequestResponse
    | APIUnauthorizedResponse
    | APIForbiddenResponse
    | APINotFoundResponse
    | APIInternalServerErrorResponse
    | APINotImplementedResponse;

alias APIBaseUpdateResponse<T> =
    | T
    | APIBadRequestResponse
    | APIUnauthorizedResponse
    | APIForbiddenResponse
    | APINotFoundResponse
    | APIInternalServerErrorResponse
    | APINotImplementedResponse;

alias APIBaseDeleteResponse<T> =
    | T
    | APIBadRequestResponse
    | APIUnauthorizedResponse
    | APIForbiddenResponse
    | APINotFoundResponse
    | APIInternalServerErrorResponse
    | APINotImplementedResponse;

alias APIResponseOrAvailableErrors<T> = APIBaseResponseOrAvailableErrors<APIResponse<T>>;
alias APICreateResponse<T> = APIBaseCreateResponse<APIResponse<T>>;
alias APIListResponse<T> = APIBaseListResponse<APIResponse<T[]>>;
alias APIGetResponse<T> = APIBaseGetResponse<APIResponse<T>>;
alias APIUpdateResponse<T> = APIBaseUpdateResponse<APIResponse<T>>;
alias APIDeleteResponse = APIBaseDeleteResponse<APINoDataResponse>;
alias APIPagedListResponse<T> = APIBaseListResponse<APIPagedResponse<T[]>>;

alias APINoDataResponseOrAvailableErrors = APIBaseResponseOrAvailableErrors<APINoDataResponse>;
alias APINoDataCreateResponse = APIBaseCreateResponse<APIResponse<int32>>;
alias APINoDataGetResponse = APIBaseGetResponse<APINoDataResponse>;
alias APINoDataListResponse = APIBaseListResponse<APINoDataResponse>;
alias APINoDataUpdateResponse = APIBaseUpdateResponse<APINoDataResponse>;
alias APINoDataDeleteResponse = APIBaseDeleteResponse<APINoDataResponse>;

model SSEBaseResponse<body> {
    @statusCode statusCode: 200;
    @header `Content-Type`: "text/event-stream";
    @header `Cache-Control`: "no-cache";
    @header Connection: "keep-alive";
    @header `Transfer-Encoding`: "chunked";
    @body A: "event:message\ndata:${body}";
}

alias SSEResponse<body> =
    | SSEBaseResponse<body>
    | APIBadRequestResponse
    | APIUnauthorizedResponse
    | APIForbiddenResponse
    | APINotFoundResponse
    | APIInternalServerErrorResponse;

model DownloadFileBaseResponse {
    @statusCode statusCode: 200;
    @header `Content-Type`: "application/octet-stream";
    @header `Content-Disposition`: "attachment; filename=file";
    @body body: bytes;
}

alias DownloadFileResponse =
    | DownloadFileBaseResponse
    | APIBadRequestResponse
    | APIUnauthorizedResponse
    | APIForbiddenResponse
    | APINotFoundResponse
    | APIInternalServerErrorResponse;
