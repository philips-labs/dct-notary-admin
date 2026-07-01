import { FC, cloneElement, MouseEvent } from 'react';
import { matchPath, useLocation, useNavigate } from 'react-router-dom';
import classNames from 'classnames';

export interface IconButton {
  label: string;
  icon: any;
}

export interface RoutedButtonProps {
  path: string;
}

export const NavButton: FC<RoutedButtonProps & IconButton> = ({ path, label, icon }) => {
  const location = useLocation();
  const navigate = useNavigate();

  const onClick = (event: MouseEvent<HTMLButtonElement>) => {
    event.preventDefault();
    navigate(path);
  };
  const pathMatch = matchPath({ path, end: true }, location.pathname);

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
