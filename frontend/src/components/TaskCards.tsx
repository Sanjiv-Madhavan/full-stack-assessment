import { useMemo } from 'react';
import { useQuery } from '@tanstack/react-query';
import {
    Alert,
    Badge,
    Box,
    Button,
    Card,
    Center,
    Group,
    Loader,
    SimpleGrid,
    Text,
    Title,
} from '@mantine/core';
import { IconAlertCircle, IconRefresh } from '@tabler/icons-react';
import { fetchAllTasks, Task, TaskStatus } from '../api/tasks';
import classes from './TaskCards.modules.css'

const statusColors: Record<TaskStatus, string> = {
    TODO: 'gray',
    IN_PROGRESS: 'yellow',
    DONE: 'green',
};

const statusLabels: Record<TaskStatus, string> = {
    TODO: 'To do',
    IN_PROGRESS: 'In progress',
    DONE: 'Done',
};

const formatDate = (value: string) => {
    const parsed = new Date(value);
    return Number.isNaN(parsed.getTime()) ? value : parsed.toLocaleString();
};

export function TaskCards() {
    const {
        data,
        isLoading,
        isError,
        error,
        refetch,
        isFetching,
    } = useQuery<Task[]>({
        queryKey: ['tasks'],
        queryFn: fetchAllTasks,
    });

    const tasks = data ?? [];

    const errorMessage = useMemo(() => {
        if (!error) {
            return 'Failed to load tasks.';
        }

        return error instanceof Error ? error.message : 'Failed to load tasks.';
    }, [error]);

    return (
        <Box className={classes.wrapper}>
            <div className={classes.heading}>
                <Title order={2}>Tasks</Title>
                <Button
                    size="xs"
                    variant="light"
                    leftSection={<IconRefresh size={14} />}
                    onClick={() => refetch()}
                    loading={isFetching}
                    disabled={isLoading}
                >
                    Refresh
                </Button>
            </div>

            {isLoading ? (
                <Center py="xl">
                    <Loader />
                </Center>
            ) : isError ? (
                <Alert
                    color="red"
                    icon={<IconAlertCircle size={16} />}
                    title="Something went wrong"
                    radius="md"
                >
                    <Text size="sm">{errorMessage}</Text>
                    <Button
                        mt="md"
                        size="xs"
                        variant="light"
                        leftSection={<IconRefresh size={14} />}
                        onClick={() => refetch()}
                    >
                        Try again
                    </Button>
                </Alert>
            ) : tasks.length === 0 ? (
                <Card withBorder radius="md" padding="xl">
                    <Text c="dimmed" ta="center">
                        There are no tasks yet. Create a project task to get started.
                    </Text>
                </Card>
            ) : (
                <SimpleGrid cols={{ base: 1, sm: 2, lg: 3 }} spacing="lg" className={classes.grid}>
                    {tasks.map((task) => (
                        <Card key={task.id} withBorder radius="md" padding="lg" className={classes.card}>
                            <Group justify="space-between" align="flex-start">
                                <Box style={{ flex: 1, marginRight: '0.75rem' }}>
                                    <Text fw={600}>{task.title}</Text>
                                    {task.description ? (
                                        <Text size="sm" c="dimmed" mt="xs">
                                            {task.description}
                                        </Text>
                                    ) : null}
                                </Box>
                                <Badge color={statusColors[task.status]} variant="light">
                                    {statusLabels[task.status]}
                                </Badge>
                            </Group>
                            <div className={classes.meta}>
                                <span>Project ID: {task.projectId}</span>
                                <span>Updated: {formatDate(task.updatedAt)}</span>
                            </div>
                        </Card>
                    ))}
                </SimpleGrid>
            )}
        </Box>
    );
}