import { FC } from 'react';
import { Box, Text } from 'grommet';
import { Info, Alert } from 'grommet-icons';

interface NotificationProps {
  type: 'unknown' | 'error' | 'info';
  message: string;
}

export const Notification: FC<NotificationProps> = ({ type, message }) => {
  return (
    <Box
      background={type === 'error' ? 'status-error' : 'status-unknown'}
      pad="small"
      animation="fadeIn"
    >
      <Box direction="row" gap="small">
        {type === 'error' ? <Alert /> : <Info />}
        <Text>{message}</Text>
      </Box>
    </Box>
  );
};
