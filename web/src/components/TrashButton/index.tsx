import { FC } from 'react';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { faTrashAlt } from '@fortawesome/free-solid-svg-icons';

interface TrashButtonProps {
  action: () => Promise<void>;
}

export const TrashButton: FC<TrashButtonProps> = ({ action }) => {
  return (
    <button className="p-1" onClick={action}>
      <FontAwesomeIcon icon={faTrashAlt} className="text-red-600" />
    </button>
  );
};
