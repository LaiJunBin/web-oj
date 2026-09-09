import { apiClient } from '@/api';

/**
 * 從產生的 endpoint 定義取出路徑,而不是在服務層另外硬寫一次網址——路徑改了
 * 就會在 build 時因為 alias 找不到而爆,不會靜靜地打到 404。
 */
export const endpointPathByAlias = (api: readonly { alias?: string; path: string }[], alias: string) => {
	const endpoint = api.find((item) => item.alias === alias);
	if (!endpoint) {
		throw new Error('Endpoint alias not found: ' + alias);
	}
	return endpoint.path;
};

/**
 * 取下載檔名:優先用後端 Content-Disposition 給的名字(它才知道匯出檔裡是什麼、
 * 也才有正確的中文檔名),取不到或格式壞掉時才退回呼叫端給的 fallback。
 */
export const downloadFilenameFrom = (contentDisposition: string | undefined, fallback: string) => {
	if (!contentDisposition) return fallback;

	const utf8Match = contentDisposition.match(/filename\*\s*=\s*UTF-8''([^;]+)/i);
	if (utf8Match?.[1]) {
		try {
			return decodeURIComponent(utf8Match[1].trim());
		} catch {
			// 壞掉的 URI 就當作沒有,往下用一般的 filename
		}
	}

	const asciiMatch = contentDisposition.match(/filename\s*=\s*"([^"]+)"|filename\s*=\s*([^;]+)/i);
	if (asciiMatch) {
		const filename = (asciiMatch[1] ?? asciiMatch[2] ?? '').trim();
		if (filename) return filename;
	}

	return fallback;
};

/**
 * 以 blob 取得一個 GET 端點的檔案並觸發瀏覽器下載。
 *
 * 這些匯出端點在 spec 裡的回應是 void(檔案串流),typed client 幫不上忙,
 * 所以直接走底層 axios;錯誤時後端回的是 JSON,但 responseType 是 blob,
 * 得先把它讀回文字才拿得到 message——否則使用者只會看到「匯出失敗」。
 */
export const downloadBlob = async (
	path: string,
	fallbackFilename: string,
	options: { params?: Record<string, string>; failureMessage: string }
) => {
	try {
		const response = await apiClient.axios.get(path, {
			responseType: 'blob',
			withCredentials: true,
			params: options.params
		});

		const filename = downloadFilenameFrom(
			response.headers['content-disposition'],
			fallbackFilename
		);
		const blobURL = URL.createObjectURL(response.data);
		const anchor = document.createElement('a');
		anchor.href = blobURL;
		anchor.download = filename;
		anchor.click();
		setTimeout(() => URL.revokeObjectURL(blobURL), 0);
	} catch (error) {
		throw new Error(await failureReason(error, options.failureMessage));
	}
};

/**
 * 匯入失敗時的訊息:後端會把「為什麼不能匯入」寫在 message,把逐項的影響或
 * 缺漏(例如缺少的映像 tag)寫在 details,兩者都要讓使用者看到。
 */
export const importFailureMessage = (error: unknown, fallback: string) => {
	const data = (error as { response?: { data?: { message?: string; details?: string[] } } })
		?.response?.data;
	if (!data?.message) {
		return error instanceof Error && error.message ? error.message : fallback;
	}
	return data.details?.length ? `${data.message}\n${data.details.join('\n')}` : data.message;
};

/**
 * 從一個失敗的 blob 請求裡挖出後端的錯誤訊息。
 *
 * responseType: 'blob' 會讓錯誤回應的 JSON 也變成 Blob,直接讀 data.message
 * 會得到 undefined。details 也一併帶出來——匯出 archive 時後端就是用它列出
 * 「還沒建置的映像」,那份清單正是使用者要看的東西。
 */
const failureReason = async (error: unknown, fallback: string) => {
	const response = (error as { response?: { data?: unknown } })?.response;
	if (!response?.data) return fallback;

	let payload = response.data;
	if (payload instanceof Blob) {
		try {
			payload = JSON.parse(await payload.text());
		} catch {
			return fallback;
		}
	}

	const { message, details } = (payload ?? {}) as { message?: string; details?: string[] };
	if (!message) return fallback;
	return details?.length ? `${message}\n${details.join('\n')}` : message;
};
