import { FC } from 'react';
import { Info, Alert } from 'grommet-icons';
import cn from 'classnames';

interface NotificationProps {
  type: 'unknown' | 'error' | 'info';
  message: string;
}

export const Notification: FC<NotificationProps> = ({ type, message }) => {
  return (
    <div
      className={cn('text-white p-5', {
        'bg-red-500': type === 'error',
        'bg-blue-500': type === 'info',
      })}
    >
      <div className="flex flex-row">
        {type === 'error' ? <Alert className="text-white" /> : <Info />}
        <p>{message}</p>
      </div>
    </div>
  );
};
