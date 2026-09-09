<script lang="ts">
	import Swal from 'sweetalert2';
	import {
		Navbar,
		NavBrand,
		NavUl,
		NavHamburger,
		Button,
		DarkMode,
		Dropdown,
		DropdownItem,
		Modal,
		Tooltip,
		Label,
		Select,
		Fileupload,
		Checkbox,
		Input
	} from 'flowbite-svelte';
	import {
		BellActiveAltSolid,
		BellSolid,
		ChevronDownOutline,
		ClockSolid
	} from 'flowbite-svelte-icons';
	import NavLi from './NavLi.svelte';
	import { page } from '$app/state';
	import { useUserApp, useUserProfile } from '@/apps/contest/application/user';
	import { useContestStorage } from '@/apps/contest/domain/contest';
	import { useContestApp } from '@/apps/contest/application/contest';
	import { useProjectApp } from '@/apps/contest/application/project';
	import { useSettingStorage } from '@/apps/contest/domain/setting';
	import { isUserProject, useProjectStorage } from '@/apps/contest/domain/project';
	import { useTimeStorage } from '@/apps/contest/domain/time';
	import { goto, invalidateAll } from '$app/navigation';
	import { onMount } from 'svelte';
	import { browser } from '$app/environment';

	const userApp = useUserApp();
	const contestApp = useContestApp();
	const projectApp = useProjectApp();
	const userProfile = useUserProfile();

	const settingStorage = useSettingStorage();
	const contestStorage = useContestStorage();
	const projectStorage = useProjectStorage();
	const timeStorage = useTimeStorage();

	const project = $derived(projectStorage.get());
	// Submitting requires being a participant of the contest that owns this project.
	// isParticipant is only returned by the user endpoint (getUserProject), which is
	// only called when the contestant role is present, so isParticipant === true
	// already implies the contestant role — this single condition is sufficient.
	const canSubmit = $derived(!!project && isUserProject(project) && project.isParticipant);

	// --- Notifications ---
	// Browser permission is the source of truth for "can we send"; enableNotification
	// is the user's in-app on/off preference. Effective ON = supported + granted + enabled.
	const notificationSupported = browser && typeof Notification !== 'undefined';
	let notificationPermission = $state<NotificationPermission | null>(null);
	onMount(() => {
		if (notificationSupported) notificationPermission = Notification.permission;
	});
	const notificationOn = $derived(
		notificationSupported &&
			notificationPermission === 'granted' &&
			!!settingStorage.get()?.enableNotification
	);
	const setNotificationEnabled = (enabled: boolean) =>
		settingStorage.set({ ...settingStorage.get(), enableNotification: enabled });
	const enableNotifications = () => {
		if (!notificationSupported) return;
		if (notificationPermission === 'granted') {
			setNotificationEnabled(true);
			return;
		}
		if (notificationPermission === 'default') {
			Notification.requestPermission().then((permission) => {
				notificationPermission = permission;
				if (permission === 'granted') setNotificationEnabled(true);
			});
		}
	};

	let openDropdown = $state(false);
	let openProjectDropdown = $state(false);
	let openSubmissionModal = $state(false);

	// 提交冷卻:後端以 429 回應並附上 retryAfterSeconds(保證 ≥ 1)。剩餘秒數由
	// 截止時刻推算而非逐秒遞減,分頁被瀏覽器節流時才不會愈算愈慢。冷卻是逐競賽
	// 專案獨立的,切換專案時歸零。
	let cooldownEndsAt = $state(0);
	let cooldownRemaining = $state(0);
	let cooldownTimer: ReturnType<typeof setInterval> | null = null;

	const tickCooldown = () => {
		cooldownRemaining = Math.max(0, Math.ceil((cooldownEndsAt - Date.now()) / 1000));
		if (cooldownRemaining === 0 && cooldownTimer !== null) {
			clearInterval(cooldownTimer);
			cooldownTimer = null;
		}
	};

	const startCooldown = (seconds: number) => {
		cooldownEndsAt = Date.now() + seconds * 1000;
		tickCooldown();
		if (cooldownTimer === null) cooldownTimer = setInterval(tickCooldown, 1000);
	};

	const stopCooldown = () => {
		cooldownEndsAt = 0;
		cooldownRemaining = 0;
		if (cooldownTimer !== null) {
			clearInterval(cooldownTimer);
			cooldownTimer = null;
		}
	};

	// 換題就丟掉上一題的倒數,否則會在不受限的題目上看到別題的冷卻。
	$effect(() => {
		project?.id;
		stopCooldown();
	});

	$effect(() => () => stopCooldown());

	let waitContestStart = $derived.by(() =>
		timeStorage.get() !== null ? timeStorage.get()!.isInWaitContestStart() : true
	);
	$effect(() => {
		if (!project) return;
		if (!waitContestStart && project.timeLeft < 0) {
			setTimeout(() => {
				projectApp.withLoading().fetchProject(project.id);
			}, 1000);
		}
	});
</script>

<Navbar class="border-b dark:border-gray-700 dark:bg-gray-800">
	<div class="flex items-center space-x-4">
		<NavBrand href="/">
			<span class="self-center text-xl font-semibold whitespace-nowrap dark:text-white">
				Web Judge
			</span>
		</NavBrand>
		<a
			class={[
				'cursor-pointer rounded-md p-2 text-sm hover:text-blue-800 dark:text-gray-400 dark:hover:text-white',
				{ ' text-blue-800 dark:text-white': page.url.pathname === '/manual' }
			]}
			href="/manual"
			>操作手冊
		</a>

		{#if userProfile.isAdmin}
			<a
				class={[
					'cursor-pointer rounded-md p-2 text-sm hover:text-blue-800 dark:text-gray-400 dark:hover:text-white',
					{ ' text-blue-800 dark:text-white': page.url.pathname.startsWith('/dashboard') }
				]}
				href="/dashboard"
				>後台管理
			</a>
		{/if}
	</div>

	<div class="flex items-center space-x-4 lg:order-2">
		<div
			id="dropdown-trigger"
			class="cursor-pointer rounded-md p-2 text-sm hover:text-blue-800 dark:text-gray-400 dark:hover:text-white"
		>
			{contestStorage.get()?.name ?? '競賽選擇'}
			<ChevronDownOutline class="ms-2 inline h-6 w-6 dark:text-white" />
		</div>
		{#if userProfile.isLoggedIn}
			<Button
				size="sm"
				color="red"
				class="cursor-pointer"
				onclick={() => {
					userApp
						.withLoading()
						.logout()
						.then(async () => {
							await goto('/login');
							invalidateAll();
						})
						.catch(() => {
							goto('/login');
						});
				}}>登出</Button
			>
		{:else}
			<Button size="sm" color="blue" href="/login">登入</Button>
		{/if}
		<div class="flex [&>button]:p-1.5">
			{#if userProfile.isLoggedIn}
				{#if !notificationSupported || notificationPermission === 'denied'}
					<Button
						size="sm"
						color="alternative"
						class="cursor-not-allowed border-0 opacity-50"
						onclick={() => {
							if (notificationPermission === 'denied') {
								Swal.fire({
									title: '通知已被瀏覽器封鎖',
									text: '請至瀏覽器設定開啟本網站的通知權限',
									icon: 'warning'
								});
							}
						}}
					>
						<BellSolid />
					</Button>
					<Tooltip placement="bottom">
						{!notificationSupported
							? '此瀏覽器不支援通知'
							: '通知已被瀏覽器封鎖，請至瀏覽器設定開啟'}
					</Tooltip>
				{:else if notificationOn}
					<Button
						size="sm"
						color="alternative"
						class="cursor-pointer border-0"
						onclick={() => setNotificationEnabled(false)}
					>
						<BellActiveAltSolid />
					</Button>
					<Tooltip placement="bottom">關閉通知</Tooltip>
				{:else}
					<Button
						size="sm"
						color="alternative"
						class="cursor-pointer border-0"
						onclick={enableNotifications}
					>
						<BellSolid />
					</Button>
					<Tooltip placement="bottom">開啟通知</Tooltip>
				{/if}
			{/if}
			<DarkMode class="cursor-pointer" />
		</div>
		<NavHamburger class="m-0 me-3 md:block lg:hidden"></NavHamburger>
		<Dropdown class="z-20 w-44" triggeredBy="#dropdown-trigger" bind:open={openDropdown}>
			{#each contestApp.getCurrentAvailableContests() as { id, name }}
				<DropdownItem
					class="cursor-pointer"
					onclick={() => {
						openDropdown = false;
						contestApp.withLoading().fetchContest(id);
					}}>{name}</DropdownItem
				>
			{:else}
				<DropdownItem class="cursor-no-drop">無更多競賽</DropdownItem>
			{/each}
		</Dropdown>
	</div>
	<NavUl
		divClass="w-full lg:block lg:w-auto order-1"
		ulClass="flex items-center flex-col py-4 mt-4 lg:flex-row xl:space-x-8 space-x-4 rtl:space-x-reverse lg:border-0 lg:mt-0 lg:text-sm lg:font-medium xl:px-4"
	>
		<NavLi>
			<a
				class={[
					'cursor-pointer rounded-md p-2 text-sm hover:text-blue-800 dark:text-gray-400 dark:hover:text-white',
					{ ' text-blue-800 dark:text-white': page.url.pathname === '/schedule' }
				]}
				href="/schedule"
				>競賽日程
			</a>
		</NavLi>
		<NavLi>
			<span
				class="relative top-[2px] cursor-pointer rounded-md border px-4 py-2 hover:text-blue-800 dark:text-gray-400 dark:hover:text-white"
			>
				{project?.name ?? '專案選擇'}
				<ChevronDownOutline class="text-primary-800 ms-2 inline h-6 w-6 dark:text-white" /></span
			>
			<Dropdown bind:open={openProjectDropdown}>
				{#each projectApp.getCurrentContestAvailableProjects() as { id, name }}
					<DropdownItem
						class="cursor-pointer"
						onclick={() => {
							openProjectDropdown = false;
							projectApp.withLoading().fetchProject(id);
						}}>{name}</DropdownItem
					>
				{:else}
					<DropdownItem class="cursor-no-drop">無更多專案</DropdownItem>
				{/each}
			</Dropdown>
		</NavLi>
		{#if timeStorage.get()?.isInSubmitPeriod()}
			<li class="flex items-center py-2">
				<Button
					size="sm"
					color="green"
					class="cursor-pointer disabled:cursor-no-drop disabled:hover:bg-green-700 disabled:dark:hover:bg-green-600"
					disabled={!canSubmit || cooldownRemaining > 0}
					onclick={() => (openSubmissionModal = true)}
					>{cooldownRemaining > 0 ? `提交(${cooldownRemaining}s)` : '提交'}</Button
				>
				{#if !canSubmit}
					<Tooltip placement="bottom">
						<div class="p-2 text-sm text-gray-500 dark:text-gray-400">您不是參賽者</div>
					</Tooltip>
				{:else if cooldownRemaining > 0}
					<Tooltip placement="bottom">
						<div class="p-2 text-sm text-gray-500 dark:text-gray-400">
							本題設有提交間隔，{cooldownRemaining} 秒後可再次提交
						</div>
					</Tooltip>
				{/if}
			</li>
		{/if}
		<li class="flex items-center py-2">
			<ClockSolid class="h-6 w-6 dark:text-white" />
			<span class="ms-2">
				{#if timeStorage.get() === null || timeStorage.get()?.isContestEnd()}
					已結束
				{:else if timeStorage.get()?.isInWaitContestStart()}
					{timeStorage.get()?.time} 後開放提交
				{:else}
					{timeStorage.get()?.time}
				{/if}
			</span>
		</li>
	</NavUl>
</Navbar>

{#if project && isUserProject(project) && project.isParticipant}
	<Modal
		title="提交"
		bind:open={openSubmissionModal}
		autoclose
		outsideclose
		footerClass="flex justify-end items-center p-4 md:p-5 space-x-3 rtl:space-x-reverse rounded-b-lg"
	>
		<form
			id="submit-form"
			class="space-y-6"
			onsubmit={async (e) => {
				e.preventDefault();
				interface SubmitProjectFormElement extends HTMLFormControlsCollection {
					project: HTMLInputElement;
					gradingItemId: HTMLSelectElement;
				}
				const { project: projectInput, gradingItemId: gradingItemIdInput } = e.currentTarget
					.elements as unknown as SubmitProjectFormElement;

				if (projectInput.files!.length === 0) {
					Swal.fire({
						icon: 'error',
						title: '提交失敗',
						text: '請上傳專案壓縮檔'
					});
					return;
				}

				const projectFile = projectInput.files![0];
				const dynamicFields: Record<string, string | File> = {};
				for (const element of e.currentTarget.elements) {
					if (element instanceof HTMLInputElement) {
						if (element.name === 'project' || element.name === 'gradingItemId') continue;
						if (element.type === 'file') {
							if (element.files!.length === 0) {
								continue;
							}
							const fr = new FileReader();
							await new Promise<void>((resolve) => {
								fr.onload = () => {
									dynamicFields[element.name] = fr.result as string;
									resolve();
								};
								fr.readAsDataURL(element.files![0]);
							});
						} else if (element.type === 'checkbox') {
							dynamicFields[element.name] = element.checked.toString();
						} else {
							dynamicFields[element.name] = element.value;
						}
					}
				}

				const payload = {
					project: projectFile,
					gradingItemIDs:
						gradingItemIdInput.value === 'all' ? [] : [parseInt(gradingItemIdInput.value)],
					dynamicFields: JSON.stringify(dynamicFields)
				};

				projectApp
					.submitProject(project.id, payload)
					.then(() => {
						// 「已送出」而非「提交成功」:落地發生在後端回應之後的非同步流程,
						// 提交出現在列表上(該事件在寫入完成後才發布)才是真正的確認點。
						Swal.fire({
							icon: 'success',
							title: '已送出',
							text: '請於提交列表確認這筆提交已出現並開始評測'
						});
						openSubmissionModal = false;
						projectApp.fetchProject(project.id);
					})
					.catch((error) => {
						console.error(error);
						// Defense in depth: the submit button/modal are already gated on
						// isParticipant, so this 403 is an edge case (e.g. the user was removed
						// from the participant list after the project loaded). Show a clear
						// message instead of the generic "submit failed".
						if (error?.response?.status === 403) {
							Swal.fire({
								icon: 'error',
								title: '提交失敗',
								text: '您不在本場次參賽名單，無法提交'
							});
						} else if (error?.response?.status === 429) {
							// 冷卻中不是失敗,是規則正常運作:告知還要等多久,並讓提交
							// 按鈕自己倒數,而不是丟一個看起來像出錯的訊息。
							const retryAfter = Number(error.response.data?.retryAfterSeconds) || 1;
							startCooldown(retryAfter);
							Swal.fire({
								icon: 'info',
								title: '提交過於頻繁',
								text: `本題設有提交間隔，請於 ${retryAfter} 秒後再試`
							});
						} else {
							// 400 現在是「這次提交真的沒有成立」——後端改成同步驗證動態欄位,
							// 不再回 200 然後讓提交無聲消失。原因只有後端知道(哪個欄位、
							// 為什麼),所以照原文轉達;沒有原文時才退回一句通則。
							const data = error?.response?.data as
								| { message?: string; details?: string[] }
								| undefined;
							const reason = [data?.message, ...(data?.details ?? [])]
								.filter(Boolean)
								.join('\n');
							Swal.fire({
								icon: 'error',
								title: '提交失敗',
								text: reason || '請確認檔案與欄位後再試一次'
							});
						}
					});
			}}
		>
			<div>
				<Label for="grading-item-id-input">
					<span class="align-middle text-red-500">*</span>
					評分項目
				</Label>

				<Select
					id="grading-item-id-input"
					name="gradingItemId"
					required
					value="all"
					items={[
						{
							value: 'all',
							name: '全部(ALL)'
						},
						...(project
							? project!.gradingItems.map((gradingItem) => ({
									value: gradingItem.id,
									name: gradingItem.name
								}))
							: [])
					]}
				></Select>
			</div>

			<div>
				<Label for="zip-file-input">
					<span class="align-middle text-red-500">*</span>
					專案壓縮檔(.zip)
				</Label>
				<Fileupload required id="zip-file-input" name="project" accept=".zip" />
			</div>

			{#if project !== null}
				{#each project.dynamicFields as field}
					<div>
						{#if field.type === 'boolean'}
							<Label class="flex items-center" for={field.name}>
								{#if field.isRequired}
									<span class="align-middle text-red-500">*</span>
								{/if}
								<Checkbox
									id={field.name}
									name={field.name}
									required={field.isRequired}
									checked={field.defaultValue === 'true'}
									class="mr-2 ml-1"
								/>
								{field.label}
							</Label>
						{:else}
							<Label for={field.name} class="mb-2 inline-block text-neutral-700 dark:text-white">
								{#if field.isRequired}
									<span class="align-middle text-red-500">*</span>
								{/if}
								{field.label}
							</Label>
							{#if field.type === 'file'}
								<Fileupload id={field.name} name={field.name} required={field.isRequired} />
							{:else}
								<Input
									id={field.name}
									name={field.name}
									value={field.defaultValue}
									required={field.isRequired}
								/>
							{/if}
						{/if}
					</div>
				{/each}
			{/if}
		</form>

		<svelte:fragment slot="footer">
			<Button
				color="green"
				class="cursor-pointer disabled:cursor-no-drop"
				type="submit"
				form="submit-form"
				disabled={cooldownRemaining > 0}
				onclick={(e: any) => e.stopPropagation()}
				>{cooldownRemaining > 0 ? `${cooldownRemaining} 秒後可再提交` : '送出'}</Button
			>
		</svelte:fragment>
	</Modal>
{/if}
