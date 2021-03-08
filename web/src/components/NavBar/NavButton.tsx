import { FC, cloneElement, MouseEvent } from 'react';
import { matchPath, useHistory, useLocation, useRouteMatch } from 'react-router';
import classNames from 'classnames';

export interface IconButton {
  label: string;
  icon: any;
}

export interface RoutedButtonProps {
  path: string;
}

export const NavButton: FC<RoutedButtonProps & IconButton> = ({ path, label, icon }) => {
  const match = useRouteMatch(path);
  const location = useLocation();
  const history = useHistory();

  const onClick = (event: MouseEvent<HTMLButtonElement>) => {
    event.preventDefault();
    history.push(path);
  };
  const pathMatch = matchPath(location.pathname, { exact: true, path });

  return (
    <button
      className={classNames('hover:bg-blue-300 hover:text-black focus:outline-none', {
        'bg-blue-300 text-black': !!pathMatch,
        'bg-transparent text-white': !pathMatch,
      })}
      onClick={onClick}
    >
      <div className="flex flex-col p-3 items-center space-y-1 justify-center">
        {cloneElement(icon, { color: !!pathMatch ? 'black' : 'white' })}
        <span className="text-xs">{label}</span>
      </div>
    </button>
  );
};
