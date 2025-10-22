import { Box, Flex, Text, ThemeIcon, Image, Loader } from '@mantine/core';
import { useQuery } from '@tanstack/react-query';
import { getHealth } from '../api/generated';
import { IconCheck, IconX } from '@tabler/icons-react';
import favicon from '../assets/favicon.svg';

export function Header() {
  const { data: healthData, isLoading } = useQuery({
    queryKey: ['health'],
    queryFn: async () => {
      const response = await getHealth({});
      return response.data;
    },
    refetchInterval: 5000,
  });

  const isHealthy = healthData?.status === 'ok';

  return (
    <Box
      component="header"
      pos="sticky"
      top={0}
      style={{ zIndex: 1000, boxShadow: '0 1px 3px rgba(0, 0, 0, 0.12), 0 1px 2px rgba(0, 0, 0, 0.24)'}}
      bg="dark.7"
      px="md"
      py="sm"
    >
      <Flex justify="space-between" align="center">
        <Flex gap="sm" align="center">
            <Image
                src={favicon}
                alt="Logo"
                width={32}
                height={32}
                fit="contain"
            />
        </Flex>

        {isLoading ? (
          <Loader size="sm" />
        ) : (
            <Box
            style={{
              border: `1px solid ${isHealthy ? '#51cf66' : '#ff6b6b'}`,
              borderRadius: '20px',
              padding: '6px 12px',
              backgroundColor: isHealthy ? 'rgba(81, 207, 102, 0.1)' : 'rgba(255, 107, 107, 0.1)'
            }}
          >
            <Flex gap="sm" align="center">
              <ThemeIcon
                color={isHealthy ? 'green' : 'red'}
                variant="light"
                size="sm"
                radius="xl"
              >
                {isHealthy ? <IconCheck size={14} /> : <IconX size={14} />}
              </ThemeIcon>
              <Text fw={500} c={isHealthy ? 'green' : 'red'} size="sm">
                {isHealthy ? 'Healthy' : 'Unhealthy'}
              </Text>
            </Flex>
          </Box>
        )}
      </Flex>
    </Box>
  );
}
