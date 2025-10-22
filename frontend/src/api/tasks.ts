import { API_BASE_URL } from './hey-api';

export type TaskStatus = 'TODO' | 'IN_PROGRESS' | 'DONE';

export interface Task {
    id: string;
    projectId: string;
    title: string;
    description?: string | null;
    status: TaskStatus;
    createdAt: string;
    updatedAt: string;
}

interface Project {
    id: string;
    name: string;
    createdAt: string;
    updatedAt: string;
}

const toError = async (response: Response) => {
    let details = '';

    try {
        const data = await response.json();
        if (data && typeof data === 'object' && 'message' in data) {
            details = `: ${String((data as { message?: unknown }).message ?? '')}`;
        }
    } catch {
        // Ignore JSON parse errors and fall back to status text.
    }

    return new Error(`Request failed with status ${response.status}${details}`);
};

export async function fetchAllTasks(): Promise<Task[]> {
    const projectResponse = await fetch(`${API_BASE_URL}/projects`);
    if (!projectResponse.ok) {
        throw await toError(projectResponse);
    }

    const projects: Project[] = await projectResponse.json();
    if (!Array.isArray(projects) || projects.length === 0) {
        return [];
    }

    const taskLists = await Promise.all(
        projects.map(async (project) => {
            const taskResponse = await fetch(`${API_BASE_URL}/projects/${project.id}/tasks`);
            if (!taskResponse.ok) {
                throw await toError(taskResponse);
            }

            const tasks: Task[] = await taskResponse.json();
            return tasks;
        }),
    );

    return taskLists.flat();
}