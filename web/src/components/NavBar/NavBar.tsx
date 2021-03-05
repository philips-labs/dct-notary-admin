import { FC, ClassAttributes, HTMLAttributes } from 'react';
import { Home, Database } from 'grommet-icons';
import { NavButton } from './NavButton';

export const NavBar: FC<ClassAttributes<HTMLDivElement> & HTMLAttributes<HTMLDivElement>> = () => {
  return (
    <div className="bg-blue-600 flex flex-col justify-between w-20 text-white font-semibold h-screen">
      <div className="flex flex-col">
        <NavButton path="/" icon={<Home />} label="Home" />
        <NavButton path="/targets" icon={<Database />} label="Targets" />
      </div>
      <div className="content-center justify-center flex p-2 border-t-2 border-white">
        <div className="border-2 border-pink-400 rounded-full w-12 h-12 p-3 flex text-center justify-center bg-white">
          <span className="self-center tracking-tighter text-xl font-semibold text-gray-900">
            MF
          </span>
        </div>
      </div>
    </div>
  );
};
