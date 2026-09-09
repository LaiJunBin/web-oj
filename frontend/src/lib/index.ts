// place files you want to import through the `$lib` alias in this folder.

/**
 * 把 number input 的綁定值正規化成 `number | undefined`。
 *
 * number input 清空後綁定到的是空字串,而 API 對「沒有給這個欄位」與「給了 0」
 * 是兩種意思(例如 submissionCooldownSeconds:未給 = 沿用系統預設 / 更新時維持
 * 現值,0 = 明確不限制)。所以這裡不能用 truthy 判斷——那會把 0 一起吃掉。
 *
 * 回傳 `undefined` 而非 `null` 是必要的:JSON.stringify 會整個略過值為 undefined
 * 的 key,送出 `null` 則會被後端當成一個明確的值,結果是回應 200 但設定沒有生效。
 */
export function optionalNumber(value: unknown): number | undefined {
	return value === '' || value === null || value === undefined ? undefined : Number(value);
}

/**
 * UTC 時間 → `<input type="datetime-local">` 的值。
 *
 * datetime-local 承載的是沒有時區資訊的「牆鐘時間」,而送出時 `new Date(值)` 會
 * 以**瀏覽器所在時區**解析它,所以這裡必須用同一個時區換算回去,來回才對稱。
 * 用 getTimezoneOffset() 而不是寫死 +8:寫死的話只有在 UTC+8 的機器上存檔才不會
 * 位移,換個時區就會每存一次推移一次。
 *
 * @param value 省略時代表「現在」
 */
export function toDatetimeLocal(value: string | number | Date = new Date()): string {
	const date = new Date(value);
	return new Date(date.getTime() - date.getTimezoneOffset() * 60_000).toISOString().slice(0, 16);
}

export function datetimeString(
	date: Date | string | number | null | undefined,
	defaultText: string = ''
): string {
	if (!date) return defaultText;

	const dateObj = new Date(date);
	if (isNaN(dateObj.getTime())) {
		return defaultText;
	}

	const year = dateObj.getFullYear();
	const month = (dateObj.getMonth() + 1).toString().padStart(2, '0');
	const day = dateObj.getDate().toString().padStart(2, '0');
	const hour = dateObj.getHours().toString().padStart(2, '0');
	const minute = dateObj.getMinutes().toString().padStart(2, '0');
	const second = dateObj.getSeconds().toString().padStart(2, '0');

	return `${year}-${month}-${day} ${hour}:${minute}:${second}`;
}

export function fileSize(bytes: number): string {
	const sizes = ['Bytes', 'KB', 'MB', 'GB', 'TB'];
	if (bytes === 0) return '0 Byte';
	const i = parseInt(String(Math.floor(Math.log(bytes) / Math.log(1024))));
	let size = (bytes / Math.pow(1024, i)).toFixed(2);
	if (size.endsWith('.00')) {
		size = size.slice(0, -3);
	} else if (size.endsWith('0')) {
		size = size.slice(0, -1);
	}
	return size + ' ' + sizes[i];
}
