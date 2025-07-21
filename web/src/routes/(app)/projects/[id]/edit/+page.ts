import { ProjectAPI } from '$lib/api/projects';
import type { LoadEvent } from '@sveltejs/kit';
import { error } from '@sveltejs/kit';

export const load = async ({ params }: LoadEvent) => {
	try {
		const projectId = params.id;

		if (!projectId) {
			throw error(400, 'Project ID is required');
		}

		const project = await ProjectAPI.getProject(projectId);

		return {
			project
		};
	} catch (err) {
		console.error('Failed to load project for editing:', err);

		if (err && typeof err === 'object' && 'status' in err) {
			throw err;
		}

		throw error(500, 'Failed to load project data');
	}
};