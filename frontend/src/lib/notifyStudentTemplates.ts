export type NotifyTemplate = { key: string; label: string; subject: string; content: string };

/**
 * The three canned messages behind the judgment page's 「通知選手」 button.
 *
 * `submittedAt` is the submission time already formatted with datetimeString —
 * the same string the contestant sees as 「提交時間」 in their submission list —
 * so a message names a submission the contestant can actually recognise. (The
 * submission id is a Mongo hex the contestant never sees.) The admin edits
 * freely before sending; this originates a directed clarification
 * (receiverUserId set, sendToAll false), never a broadcast.
 */
export function notifyStudentTemplates(submittedAt: string): NotifyTemplate[] {
	return [
		{
			key: 'checked',
			label: '檢查有結果',
			subject: `${submittedAt} 的提交:檢查結果`,
			content: `你的 ${submittedAt} 的提交檢查已有結果,說明已附註於該筆提交,請開啟提交詳情查看。`
		},
		{
			key: 'rejudged',
			label: '已重新評測',
			subject: `${submittedAt} 的提交:已重新評測`,
			content: `你的 ${submittedAt} 的提交已重新評測,請查看最新結果。`
		},
		{
			key: 'resubmit',
			label: '需修正後重交',
			subject: `${submittedAt} 的提交:需修正後重新提交`,
			content: `你的 ${submittedAt} 的提交需修正後重新提交,原因已附註於該筆提交。`
		}
	];
}
