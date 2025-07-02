import React from 'react';
import { Bars3Icon, HomeIcon, CogIcon, ChartBarIcon } from '@heroicons/react/24/outline';

export const Navbar: React.FC = () => {
  return (
    <nav className="bg-blue-600 shadow-lg">
      <div className="container mx-auto px-4">
        <div className="flex justify-between items-center h-16">
          <div className="flex items-center">
            <div className="flex-shrink-0">
              <h1 className="text-white text-xl font-bold">NRL Scheduler</h1>
            </div>
            <div className="hidden md:block">
              <div className="ml-10 flex items-baseline space-x-4">
                <NavLink href="/" icon={<HomeIcon className="h-5 w-5" />}>
                  Dashboard
                </NavLink>
                <NavLink href="/constraints" icon={<CogIcon className="h-5 w-5" />}>
                  Constraints
                </NavLink>
                <NavLink href="/draws" icon={<ChartBarIcon className="h-5 w-5" />}>
                  Draws
                </NavLink>
              </div>
            </div>
          </div>
          <div className="md:hidden">
            <button
              type="button"
              className="text-white hover:text-gray-300 focus:outline-none focus:text-gray-300"
            >
              <Bars3Icon className="h-6 w-6" />
            </button>
          </div>
        </div>
      </div>
    </nav>
  );
};

interface NavLinkProps {
  href: string;
  icon: React.ReactNode;
  children: React.ReactNode;
}

const NavLink: React.FC<NavLinkProps> = ({ href, icon, children }) => {
  return (
    <a
      href={href}
      className="text-gray-300 hover:bg-blue-700 hover:text-white px-3 py-2 rounded-md text-sm font-medium flex items-center space-x-1"
    >
      {icon}
      <span>{children}</span>
    </a>
  );
};