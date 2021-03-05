import { FC } from 'react';
import { Button } from 'grommet';
import { Trash } from 'grommet-icons';

interface TrashButtonProps {
  action: () => Promise<void>;
}

export const TrashButton: FC<TrashButtonProps> = ({ action }) => {
  return <Button icon={<Trash color="red" />} onClick={() => action()} />;
};
