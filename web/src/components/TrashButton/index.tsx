import { FC } from 'react';
import { Trash } from 'grommet-icons';

interface TrashButtonProps {
  action: () => Promise<void>;
}

export const TrashButton: FC<TrashButtonProps> = ({ action }) => {
  return (
    <button className="p-1" onClick={action}>
      <Trash color="red" />
    </button>
  );
};
