<script lang="ts">
	import { goto, invalidateAll } from '$app/navigation';
	import Swal from 'sweetalert2';

	import { useUserApp } from '@/apps/contest/application/user';
	import { useProjectApp } from '@/apps/contest/application/project';
	import { useContestStorage } from '@/apps/contest/domain/contest';
	import { useProjectStorage } from '@/apps/contest/domain/project';
	import { useContestApp } from '@/apps/contest/application/contest';
	import { LoginFailedError } from '@/apps/contest/domain/user';

	let username = $state('');
	let password = $state('');

	const userApp = useUserApp();
	const contestApp = useContestApp();
	const projectApp = useProjectApp();
	const contestStorage = useContestStorage();
	const projectStorage = useProjectStorage();
</script>

<div
	class="mx-auto mt-8 h-fit w-full max-w-sm rounded-lg border border-gray-200 bg-white p-6 shadow dark:bg-gray-800 dark:text-white"
>
	<h1
		class="text-center text-4xl leading-none font-extrabold tracking-tight text-gray-900 dark:text-white"
	>
		登入
	</h1>
	<hr class="my-4" />
	<form
		method="POST"
		onsubmit={(e) => {
			e.preventDefault();
			userApp
				.withLoading()
				.login(username, password)
				.then(() => {
					const project = projectStorage.get();
					const contest = contestStorage.get();
					if (!project || !contest) {
						return goto('/').then(() => invalidateAll());
					}
					return Promise.all([
						projectApp.fetchProject(project.id),
						contestApp.fetchContest(contest.id)
					])
						.then(() => {
							goto('/').then(() => {
								invalidateAll();
							});
						})
						.catch(() => {
							invalidateAll();
						});
				})
				.catch((error: unknown) => {
					const status = error instanceof LoginFailedError ? error.status : undefined;
					// 429 不是帳密錯誤,是登入頻率限制。它擋的是連續失敗,所以叫人
					// 「檢查密碼」會讓他一直重試,把剩餘額度挖得更深。
					if (status === 429) {
						Swal.fire({
							title: '嘗試過於頻繁',
							text: '這不是帳號密碼錯誤。系統限制短時間內的登入次數,請稍候片刻再試一次,不需要重複點擊。',
							icon: 'warning',
							confirmButtonText: '知道了'
						});
						return;
					}
					if (status === 401) {
						Swal.fire({
							title: '登入失敗',
							text: '請檢查帳號密碼是否正確',
							icon: 'error',
							confirmButtonText: '確定'
						});
						return;
					}
					// 連不到伺服器或 5xx:講「檢查密碼」只會讓人白試,問題不在他身上。
					Swal.fire({
						title: '無法登入',
						text: '目前無法連上伺服器,請稍後再試;若持續發生請洽現場工作人員。',
						icon: 'error',
						confirmButtonText: '確定'
					});
				});
		}}
	>
		<div class="mb-6">
			<label for="username" class="mb-2 block text-sm font-medium text-gray-900 dark:text-white">
				帳號
			</label>
			<input
				type="username"
				id="username"
				name="username"
				class="block w-full rounded-lg border border-gray-300 bg-gray-50 p-2.5 text-sm text-gray-900 shadow-sm focus:border-blue-500 focus:ring-blue-500"
				placeholder="demo01"
				bind:value={username}
				required
			/>
		</div>
		<div class="mb-6">
			<label for="password" class="mb-2 block text-sm font-medium text-gray-900 dark:text-white">
				密碼
			</label>
			<input
				type="password"
				id="password"
				name="password"
				bind:value={password}
				class="block w-full rounded-lg border border-gray-300 bg-gray-50 p-2.5 text-sm text-gray-900 shadow-sm focus:border-blue-500 focus:ring-blue-500"
				required
			/>
		</div>

		<div class="text-center">
			<button
				type="submit"
				class="w-full cursor-pointer rounded-lg bg-green-700 px-5 py-2.5 text-center text-sm font-medium text-white hover:bg-green-800 focus:ring-4 focus:ring-blue-300 focus:outline-none"
			>
				登入
			</button>
		</div>
	</form>
</div>
