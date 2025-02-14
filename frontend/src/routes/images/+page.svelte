<script lang="ts">
	import { imagesStore } from '$lib/stores/docker';
	import { formatBytes } from '$lib/utils/format';
	import { dockerClient } from '$lib/api/client';

	$: images = $imagesStore.data;
	$: isLoading = $imagesStore.loading;

	// Group images by repository
	$: imagesByRepo = images.reduce(
		(acc, img) => {
			const repo = img.RepoTags?.[0]?.split(':')[0] || 'untagged';
			if (!acc[repo]) acc[repo] = [];
			acc[repo].push(img);
			return acc;
		},
		{} as Record<string, typeof images>
	);
</script>

<div class="space-y-6">
	<div class="flex items-center justify-between">
		<h1 class="text-2xl font-bold">Images</h1>
		<button
			class="rounded-lg bg-blue-500 px-4 py-2 text-white hover:bg-blue-600"
			on:click={() => imagesStore.refresh(() => dockerClient.getImages())}
		>
			Pull Image
		</button>
	</div>

	{#if isLoading}
		<div class="text-center">Loading...</div>
	{:else}
		{#each Object.entries(imagesByRepo) as [repo, images]}
			<div class="rounded-lg bg-white shadow dark:bg-gray-800">
				<div class="border-b border-gray-200 px-4 py-3 dark:border-gray-700">
					<h3 class="font-medium">{repo}</h3>
				</div>
				<div class="overflow-x-auto">
					<table class="min-w-full divide-y divide-gray-200 dark:divide-gray-700">
						<thead>
							<tr>
								<th class="px-4 py-3 text-left text-xs font-medium uppercase tracking-wider">Tag</th
								>
								<th class="px-4 py-3 text-left text-xs font-medium uppercase tracking-wider">ID</th>
								<th class="px-4 py-3 text-left text-xs font-medium uppercase tracking-wider"
									>Size</th
								>
								<th class="px-4 py-3 text-left text-xs font-medium uppercase tracking-wider"
									>Created</th
								>
								<th class="px-4 py-3 text-right text-xs font-medium uppercase tracking-wider"
									>Actions</th
								>
							</tr>
						</thead>
						<tbody class="divide-y divide-gray-200 dark:divide-gray-700">
							{#each images as image}
								<tr class="hover:bg-gray-50 dark:hover:bg-gray-700">
									<td class="whitespace-nowrap px-4 py-3">
										{image.RepoTags?.[0]?.split(':')[1] || 'latest'}
									</td>
									<td class="whitespace-nowrap px-4 py-3 font-mono">
										{image.Id.split(':')[1].substring(0, 12)}
									</td>
									<td class="whitespace-nowrap px-4 py-3">
										{formatBytes(image.Size)}
									</td>
									<td class="whitespace-nowrap px-4 py-3">
										{new Date(image.Created * 1000).toLocaleString()}
									</td>
									<td class="whitespace-nowrap px-4 py-3 text-right">
										<button
											class="text-red-600 hover:text-red-900"
											on:click={() => dockerClient.removeImage(image.Id)}
										>
											Remove
										</button>
									</td>
								</tr>
							{/each}
						</tbody>
					</table>
				</div>
			</div>
		{/each}
	{/if}
</div>
