import { useState, useRef, useEffect } from "react";
import { ImageIcon, ImagePlusIcon, LayoutDashboardIcon, ScanIcon, SearchIcon, TableIcon } from "lucide-react";
import { Form, NavLink, useSubmit } from "react-router-dom";
import { Button } from "./ui/button";
import { Input } from "./ui/input";
import { ModeToggle } from "./mode-toggle";
import { Popover, PopoverContent, PopoverTrigger } from "./ui/popover";
import { Badge } from "./ui/badge";

const navs = [
  { name: `Dashboard`, icon: <LayoutDashboardIcon className="mr-2 h-4 w-4" />, to: `/` },
  { name: `Gallery`, icon: <ImageIcon className="mr-2 h-4 w-4" />, to: `/gallery` },
  { name: `Overview`, icon: <TableIcon className="mr-2 h-4 w-4" />, to: `/overview` },
  { name: `New Probe`, icon: <ImagePlusIcon className="mr-2 h-4 w-4" />, to: `/submit` }
];

const searchOperators = [
  { key: 'title', description: 'search by title' },
  { key: 'body', description: 'search by html body' },
  { key: 'tech', description: 'search by technology' },
  { key: 'header', description: 'search by header' },
  { key: 'p', description: 'search by perception hash' },
];

const Navigation = () => {
  const [searchValue, setSearchValue] = useState("");
  const [usedOperators, setUsedOperators] = useState<string[]>([]);
  const [isPopoverOpen, setIsPopoverOpen] = useState(false);
  const inputRef = useRef<HTMLInputElement>(null);
  const submit = useSubmit();

  const handleInputChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setSearchValue(e.target.value);
    const operators = e.target.value.match(/(\w+):/g) || [];
    setUsedOperators(operators.map(op => op.slice(0, -1)));
  };

  const handleSubmit = (event: React.FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    setIsPopoverOpen(false);
    submit(event.currentTarget);
  };

  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      if (inputRef.current && !inputRef.current.contains(event.target as Node)) {
        setIsPopoverOpen(false);
      }
    };

    document.addEventListener('mousedown', handleClickOutside);
    return () => {
      document.removeEventListener('mousedown', handleClickOutside);
    };
  }, []);

  return (
    <>
      <nav className="container flex h-14 items-center">
        <NavLink to="/" className="flex items-center space-x-2 p-2">
          <ScanIcon className="h-6 w-6" />
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
        <Form method="post" action="/search" onSubmit={handleSubmit} className="ml-auto flex-1 sm:flex-initial">
          <Popover open={isPopoverOpen}>
            <PopoverTrigger asChild>
              <div className="relative">
                <SearchIcon className="absolute left-2.5 top-2.5 h-4 w-4 text-muted-foreground" />
                <Input
                  ref={inputRef}
                  name="query"
                  type="search"
                  placeholder="Search..."
                  className="pl-8 sm:w-[300px] md:w-[200px] lg:w-[300px]"
                  defaultValue={searchValue}
                  onChange={handleInputChange}
                  onFocus={() => {
                    setIsPopoverOpen(true);
                    setTimeout(() => {
                      if (inputRef.current) {
                        inputRef.current.focus();
                      }
                    }, 0);
                  }}
                />
              </div>
            </PopoverTrigger>
            <PopoverContent className="w-[300px] p-0" align="start">
              <div className="grid gap-4 p-4">
                <div className="space-y-2">
                  <h4 className="font-medium leading-none">Search Operators</h4>
                  <p className="text-sm text-muted-foreground">
                    Use these operators to refine your search.
                  </p>
                </div>
                <div className="grid gap-2">
                  {searchOperators.length === usedOperators.length && <div className="text-sm">No operators left.</div>}
                  {searchOperators.map((operator) => (
                    !usedOperators.includes(operator.key) && (
                      <div key={operator.key} className="flex items-center">
                        <Badge variant="secondary" className="mr-2">
                          {operator.key}:
                        </Badge>
                        <span className="text-sm">{operator.description}</span>
                      </div>
                    )
                  ))}
                </div>
              </div>
            </PopoverContent>
          </Popover>
        </Form>
        <ModeToggle />
      </div>
    </>
  );
};

export default Navigation;