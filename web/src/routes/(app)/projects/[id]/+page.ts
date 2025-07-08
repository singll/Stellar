import { ProjectAPI } from '$lib/api/projects';
import type { Project } from '$lib/types/project';
import type { LoadEvent } from '@sveltejs/kit';
import { error } from '@sveltejs/kit';

export const load = async ({ params }: LoadEvent) => {
  try {
    const projectId = params.id;
    
    if (!projectId) {
      throw error(400, 'Project ID is required');
    }

    // 并行获取项目详情和相关数据
    const [project, members, activities] = await Promise.allSettled([
      ProjectAPI.getProject(projectId),
      ProjectAPI.getProjectMembers(projectId),
      ProjectAPI.getProjectActivities(projectId, { page: 1, limit: 10 })
    ]);

    if (project.status !== 'fulfilled') {
      throw error(404, 'Project not found');
    }

    const projectData: Project = project.value;

    const membersData = members.status === 'fulfilled' ? members.value : [];
    const activitiesData = activities.status === 'fulfilled' ? activities.value : { data: [], total: 0 };

    return {
      project: projectData,
      members: membersData,
      activities: activitiesData,
      projectId
    };
  } catch (err) {
    console.error('Failed to load project:', err);
    
    if (err && typeof err === 'object' && 'status' in err) {
      throw err; // Re-throw SvelteKit errors
    }
    
    throw error(500, 'Failed to load project data');
  }
}; 