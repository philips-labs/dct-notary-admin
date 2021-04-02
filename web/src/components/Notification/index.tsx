import { FC } from 'react';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { faInfoCircle, faExclamationTriangle } from '@fortawesome/free-solid-svg-icons';
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
        {type === 'error' ? (
          <FontAwesomeIcon icon={faExclamationTriangle} className="text-white" />
        ) : (
          <FontAwesomeIcon icon={faInfoCircle} className="text-white" />
        )}
        <p>{message}</p>
      </div>
    </div>
  );
};
