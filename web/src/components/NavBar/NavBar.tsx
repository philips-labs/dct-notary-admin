import { FC, ClassAttributes, HTMLAttributes } from 'react';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { faDatabase } from '@fortawesome/free-solid-svg-icons';
import { NavButton } from './NavButton';

export const NavBar: FC<ClassAttributes<HTMLDivElement> & HTMLAttributes<HTMLDivElement>> = () => {
  return (
    <nav className="bg-blue-600 flex flex-col justify-between w-20 text-white font-semibold h-screen">
      <div className="flex flex-col">
        <NavButton path="/" icon={<FontAwesomeIcon icon={faDatabase} />} label="Keys" />
      </div>
      <div className="content-center justify-center flex p-2 border-t-2 border-white">
        <div className="border-2 border-pink-400 rounded-full w-12 h-12 p-3 flex text-center justify-center bg-white">
          <span className="self-center tracking-tighter text-xl font-semibold text-gray-900">
            MF
          </span>
        </div>
      </div>
    </nav>
  );
};
