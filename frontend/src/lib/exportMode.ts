import Swal from 'sweetalert2';
import type { BundleExportMode } from '@/apps/admin/domain/bundle';
import { sanitizeHtml } from '$lib/sanitize';

/**
 * 問使用者映像要以哪種形態隨行。
 *
 * 兩者差幾百 MB,而 archive 還有「所有映像都得先建置完成」的前提,選錯要整個
 * 重匯一次——所以把差別講在選擇畫面上,而不是預設一個然後讓人事後才發現。
 *
 * @returns 選定的形態;使用者取消時回 null
 */
export const askExportMode = async (title: string, intro: string): Promise<BundleExportMode | null> => {
	const result = await Swal.fire({
		title,
		html: `<div style="text-align:left">
			<p>${sanitizeHtml(intro)}</p>
			<p style="margin-top:.75rem"><b>建置原始檔</b>(小,約 KB)<br>
			匯入端要自行建置,過程需要能連外拉 base image。</p>
			<p style="margin-top:.5rem"><b>映像檔</b>(大,可能數百 MB)<br>
			匯入端免建置、免連網,適合離線考場;需要映像已建置完成。</p>
		</div>`,
		icon: 'question',
		showDenyButton: true,
		showCancelButton: true,
		confirmButtonText: '建置原始檔',
		denyButtonText: '映像檔',
		cancelButtonText: '取消'
	});

	if (result.isConfirmed) return 'source';
	if (result.isDenied) return 'archive';
	return null;
};

/** 匯出進行中的提示;回傳關閉它的函式。 */
export const showExporting = () => {
	Swal.fire({ title: '匯出中…', didOpen: () => Swal.showLoading(), allowOutsideClick: false });
	return () => Swal.close();
};

/** 匯出失敗:後端會把「還沒建置的映像」之類的細節寫在訊息裡,原文顯示。 */
export const showExportError = (error: Error) =>
	Swal.fire({
		icon: 'error',
		title: '匯出失敗',
		html: `<pre style="text-align:left;white-space:pre-wrap">${sanitizeHtml(error.message)}</pre>`
	});
