import { Container } from '@mantine/core';
import { TaskCards } from '../components/TaskCards';

export default function Dashboard() {
    return (
        <Container size="lg" py="xl">
            <TaskCards />
        </Container>
    );
}