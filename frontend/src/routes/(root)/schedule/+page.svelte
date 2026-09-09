<script lang="ts">
	import { Timeline, TimelineItem } from 'flowbite-svelte';
	import { StarSolid } from 'flowbite-svelte-icons';
	import { datetimeString } from '$lib';
	import {
		useContestScheduleStorage,
		useContestStorage,
		type ScheduleProject
	} from '@/apps/contest/domain/contest';
	import { useContestApp } from '@/apps/contest/application/contest';

	const scheduleStorage = useContestScheduleStorage();
	const contestApp = useContestApp();
	const contestStorage = useContestStorage();

	$effect(() => {
		const contestId = contestStorage.get()?.id;
		if (contestId) {
			contestApp.withLoading().fetchContestSchedule(contestId);
		}
	});

	const getContestDuration = (project: ScheduleProject) => {
		const total =
			new Date(
				new Date(project.endTime).getTime() - new Date(project.startTime).getTime()
			).getTime() / 1000;
		const day = Math.floor(total / 86400).toString();
		const hour = Math.floor((total % 86400) / 3600)
			.toString()
			.padStart(2, '0');
		const minute = Math.floor((total % 3600) / 60)
			.toString()
			.padStart(2, '0');
		const second = Math.floor(total % 60).toString().padStart(2, '0');

		if (day !== '0') {
			return `${day} 天 ${hour} 小時 ${minute} 分鐘 ${second} 秒`;
		}

		if (hour === '00') {
			if (minute === '00') {
				return `${second} 秒`;
			}
			return `${minute} 分鐘 ${second} 秒`;
		}

		return `${hour} 小時 ${minute} 分鐘 ${second} 秒`;
	};
</script>

<h1 class="mt-4 text-center text-3xl font-bold text-gray-900 dark:text-gray-100">競賽日程</h1>
<hr class="border-primary-500 dark:border-primary-400 mx-auto my-4 w-20 rounded-full border-2" />
<main class="mt-4 flex justify-center">
	<Timeline order="vertical">
		{#each scheduleStorage.get()?.projects ?? [] as project}
			<TimelineItem
				classTime="text-lg"
				classH3="text-2xl mb-2"
				classDiv="mt-1"
				title={`${project.name} (${project.score}分) (競賽時間: ${getContestDuration(project)})`}
				date={`${datetimeString(project.startTime) || ''} - ${datetimeString(project.endTime) || ''}`}
			>
				<svelte:fragment slot="icon">
					<span
						class="bg-primary-200 dark:bg-primary-900 absolute -start-3 mt-1 flex h-6 w-6 items-center justify-center rounded-full ring-8 ring-white dark:ring-gray-900"
					>
						<StarSolid class="h-4 w-4" />
					</span>
				</svelte:fragment>
				<p class="mb-4 text-lg font-normal text-gray-500 dark:text-gray-400">
					{project.description}
				</p>
			</TimelineItem>
		{:else}
			<p class="text-center text-2xl text-gray-500 dark:text-gray-400">尚未有競賽日程</p>
		{/each}
	</Timeline>
</main>
