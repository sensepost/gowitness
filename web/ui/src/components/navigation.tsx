import {
  HomeIcon,
  ImageIcon,
  MonitorDownIcon,
  PlusCircleIcon,
  SearchIcon,
  TableIcon
} from "lucide-react";
import { NavLink } from "react-router-dom";
import { Button } from "./ui/button";
import { Input } from "./ui/input";
import { ModeToggle } from "./mode-toggle";

const navs = [
  { name: `Home`, icon: <HomeIcon className="mr-2 h-4 w-4" />, to: `/` },
  { name: `Gallery`, icon: <ImageIcon className="mr-2 h-4 w-4" />, to: `/gallery` },
  { name: `Table`, icon: <TableIcon className="mr-2 h-4 w-4" />, to: `/table` },
  { name: `Submit`, icon: <PlusCircleIcon className="mr-2 h-4 w-4" />, to: `/submit` },
];

const Navigation = () => {
  return (
    <>
      <nav className="container flex h-14 items-center">
        <NavLink to="/" className="flex items-center space-x-2">
          <MonitorDownIcon className="h-6 w-6" />
          <span className="font-bold">gowitness, v3</span>
        </NavLink>
        {navs.map(nav => {
          return <NavLink
            key={nav.to}
            to={nav.to}
            className={({ isActive }) =>
              isActive
                ? "text-foreground transition-colors hover:text-foreground"
                : "text-muted-foreground transition-colors hover:text-foreground"
            }
          >
            <Button variant="ghost" size="default">
              {nav.icon} {nav.name}
            </Button>
          </NavLink>;
        })}
      </nav>

      <div className="flex w-full items-center gap-4 md:ml-auto md:gap-2 lg:gap-4">
        <form className="ml-auto flex-1 sm:flex-initial">
          <div className="relative">
            <SearchIcon className="absolute left-2.5 top-2.5 h-4 w-4 text-muted-foreground" />
            <Input
              type="search"
              placeholder="Search..."
              className="pl-8 sm:w-[300px] md:w-[200px] lg:w-[300px]"
            />
          </div>
        </form>
        <ModeToggle />
      </div>
    </>
  );
};

export default Navigation;