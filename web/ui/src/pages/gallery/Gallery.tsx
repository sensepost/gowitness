import { Button } from "@/components/ui/button";
import { Card, CardContent, CardFooter } from "@/components/ui/card";
import { useEffect, useMemo, useState } from "react";
import { Link, useSearchParams } from "react-router-dom";
import { WideSkeleton } from "@/components/loading";
import { Badge } from "@/components/ui/badge";
import {
  AlertOctagonIcon, BanIcon, CheckIcon, ChevronLeftIcon, ChevronRightIcon, ClockIcon, ExternalLinkIcon,
  FilterIcon, GroupIcon, ShieldCheckIcon, XIcon
} from "lucide-react";
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from "@/components/ui/tooltip";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Popover, PopoverContent, PopoverTrigger } from "@/components/ui/popover";
import { Command, CommandEmpty, CommandGroup, CommandInput, CommandItem, CommandList } from "@/components/ui/command";
import { formatDistanceToNow, format } from 'date-fns';
import { cn } from "@/lib/utils";
import * as api from "@/lib/api/api";
import * as apitypes from "@/lib/api/types";
import { getData, getWappalyzerData } from "./data";
import { getIconUrl, getStatusColor } from "@/lib/common";
import { Label } from "@/components/ui/label";
import { Switch } from "@/components/ui/switch";


const GalleryPage = () => {
  const [gallery, setGallery] = useState<apitypes.galleryResult[]>();
  const [wappalyzer, setWappalyzer] = useState<apitypes.wappalyzer>();
  const [technology, setTechnology] = useState<apitypes.technologylist>();
  const [totalPages, setTotalPages] = useState(0);
  const [loading, setLoading] = useState(true);

  const [searchParams, setSearchParams] = useSearchParams();
  // pagination
  const page = parseInt(searchParams.get("page") || "1");
  const limit = parseInt(searchParams.get("limit") || "24");
  //filters
  const technologyFilter = searchParams.get("technologies") || "";
  const statusFilter = searchParams.get("status") || "";
  // toggles
  const perceptionGroup = searchParams.get("perception") === "true";
  const showFailed = searchParams.get("failed") !== "false"; // Default to true

  useEffect(() => {
    getWappalyzerData(setWappalyzer, setTechnology);
  }, []);

  useEffect(() => {
    getData(
      setLoading, setGallery, setTotalPages,
      page, limit, technologyFilter, statusFilter, perceptionGroup, showFailed
    );
  }, [page, limit, perceptionGroup, statusFilter, technologyFilter, showFailed]);

  const handlePageChange = (newPage: number) => {
    setSearchParams(prev => {
      prev.set("page", newPage.toString());
      return prev;
    });
  };

  const handleLimitChange = (newLimit: string) => {
    setSearchParams(prev => {
      prev.set("limit", newLimit);
      return prev;
    });
  };

  const handleTechnologyChange = (tech: string) => {
    const field = "technologies";
    setSearchParams(prev => {
      const currentTechnology = prev.get(field)?.split(",").filter(Boolean) || [];

      if (currentTechnology.includes(tech)) {
        const updatedTechnology = currentTechnology.filter(s => s !== tech);
        prev.set(field, updatedTechnology.join(","));
      } else {
        currentTechnology.push(tech);
        prev.set(field, currentTechnology.join(","));
      }

      return prev;
    });
    handlePageChange(1); // back to page 1
  };

  const handleStatusFilter = (status: string) => {
    setSearchParams(prev => {
      const currentStatus = prev.get("status")?.split(",").filter(Boolean) || [];

      if (currentStatus.includes(status)) {
        const updatedStatus = currentStatus.filter(s => s !== status);
        prev.set("status", updatedStatus.join(","));
      } else {
        currentStatus.push(status);
        prev.set("status", currentStatus.join(","));
      }

      return prev;
    });
  };

  const handleGroupBySimilar = () => {
    setSearchParams(prev => {
      prev.set("perception", (!perceptionGroup).toString());
      return prev;
    });
  };

  const handleToggleShowFailed = () => {
    setSearchParams(prev => {
      prev.set("failed", (!showFailed).toString());
      return prev;
    });
  };

  const sortedTechnologies = useMemo(() => {
    if (!technology) return [];
    const selectedTechnologies = technologyFilter.split(',').filter(Boolean);
    return [
      ...selectedTechnologies,
      ...technology.technologies.filter(tech => !selectedTechnologies.includes(tech))
    ];
  }, [technology, technologyFilter]);

  const renderPageButtons = (visible: number) => {
    const pageButtons = [];
    const maxVisiblePages = visible;
    const startPage = Math.max(1, page - Math.floor(maxVisiblePages / 2));
    const endPage = Math.min(totalPages, startPage + maxVisiblePages - 1);

    for (let i = startPage; i <= endPage; i++) {
      pageButtons.push(
        <Button
          key={i}
          onClick={() => handlePageChange(i)}
          variant={i === page ? "secondary" : "outline"}
          size="sm"
        >
          {i}
        </Button>
      );
    }

    return pageButtons;
  };

  const renderGalleryCard = (screenshot: apitypes.galleryResult) => {
    const probedDate = new Date(screenshot.probed_at);
    const timeAgo = formatDistanceToNow(probedDate, { addSuffix: true });
    const rawDate = format(probedDate, "PPpp"); // Formats the date in a readable format

    return (
      <Link to={`/screenshot/${screenshot.id}`} key={screenshot.id}>
        <Card className="group overflow-hidden transition-all hover:shadow-lg flex flex-col h-full">
          <CardContent className="p-0 relative flex-grow">
            {screenshot.failed ? (
              <div className="w-full h-48 bg-gray-800 flex items-center justify-center">
                <XIcon className="text-gray-600 w-12 h-12" />
              </div>
            ) : (
              <img
                src={screenshot.screenshot
                  ? `data:image/png;base64,${screenshot.screenshot}`
                  : api.endpoints.screenshot.path + "/" + screenshot.file_name}
                alt={screenshot.url}
                loading="lazy"
                className="w-full h-48 object-cover transition-all duration-300 filter group-hover:scale-105"
              />
            )}
            <div className="absolute top-2 right-2">
              <Badge variant="default" className={`${getStatusColor(screenshot.response_code)}`}>
                {screenshot.response_code}
              </Badge>
            </div>
            <div className="absolute bottom-2 right-2 opacity-0 group-hover:opacity-100 transition-opacity">
              <ExternalLinkIcon className="text-white drop-shadow-lg" />
            </div>
          </CardContent>

          <CardFooter className="p-2 flex flex-col items-start">
            <div className="w-full mb-2">
              <TooltipProvider>
                <Tooltip>
                  <TooltipTrigger asChild>
                    <div className="w-full truncate text-sm font-medium">
                      {screenshot.title || "Untitled"}
                    </div>
                  </TooltipTrigger>
                  <TooltipContent>
                    <p>{screenshot.title || "Untitled"}</p>
                  </TooltipContent>
                </Tooltip>
              </TooltipProvider>
              <div className="w-full truncate text-xs text-muted-foreground mt-1">
                {screenshot.url}
              </div>
            </div>
            <div className="w-full flex items-center justify-between mt-2">
              <TooltipProvider delayDuration={0}>
                <Tooltip>
                  <TooltipTrigger asChild>
                    <div className="flex items-center space-x-1 text-xs text-muted-foreground">
                      <ClockIcon className="w-3 h-3" />
                      <span className="text-nowrap">{timeAgo}</span>
                    </div>
                  </TooltipTrigger>
                  <TooltipContent side="bottom" className="text-xs">
                    <p>{rawDate}</p>
                  </TooltipContent>
                </Tooltip>
              </TooltipProvider>
              <div className="flex flex-wrap justify-end gap-1">
                {screenshot.technologies?.map(tech => {
                  const iconUrl = getIconUrl(tech, wappalyzer);
                  return iconUrl ? (
                    <TooltipProvider key={tech} delayDuration={0}>
                      <Tooltip>
                        <TooltipTrigger asChild>
                          <div className="w-6 h-6 flex items-center justify-center">
                            <img
                              src={iconUrl}
                              alt={tech}
                              loading="lazy"
                              className="w-5 h-5 object-contain"
                            />
                          </div>
                        </TooltipTrigger>
                        <TooltipContent>
                          <p>{tech}</p>
                        </TooltipContent>
                      </Tooltip>
                    </TooltipProvider>
                  ) : null;
                })}
              </div>
            </div>
          </CardFooter>
        </Card>
      </Link>
    );
  };



  if (loading) return <WideSkeleton />;

  return (
    <div className="space-y-6">
      <div className="flex flex-wrap gap-4 items-center justify-between rounded-lg">
        <div className="flex flex-wrap gap-2">
          <Popover>
            <PopoverTrigger asChild>
              <Button variant="outline" className="w-[200px] justify-start">
                <FilterIcon className="mr-2 h-4 w-4" />
                {technologyFilter.split(',').filter(n => n).length > 0 ? (
                  <>
                    {technologyFilter.split(',').length} selected
                  </>
                ) : (
                  "Filter by Technology"
                )}
              </Button>
            </PopoverTrigger>
            <PopoverContent className="w-[200px] p-0">
              <Command>
                <CommandInput placeholder="Search technologies..." />
                <CommandList>
                  <CommandEmpty>No technology found.</CommandEmpty>
                  <CommandGroup>
                    {sortedTechnologies.map((tech) => (
                      <CommandItem
                        key={tech}
                        onSelect={() => handleTechnologyChange(tech)}
                      >
                        <CheckIcon
                          className={cn(
                            "mr-2 h-4 w-4",
                            technologyFilter.includes(tech) ? "opacity-100" : "opacity-0"
                          )}
                        />
                        {tech}
                      </CommandItem>
                    ))}
                  </CommandGroup>
                </CommandList>
              </Command>
            </PopoverContent>
          </Popover>
          <Button
            variant={statusFilter.includes("200") ? "secondary" : "outline"}
            onClick={() => handleStatusFilter("200")}
          >
            <ShieldCheckIcon className="mr-2 h-4 w-4" />
            200
          </Button>
          <Button
            variant={statusFilter.includes("403") ? "secondary" : "outline"}
            onClick={() => handleStatusFilter("403")}
          >
            <BanIcon className="mr-2 h-4 w-4" />
            403
          </Button>
          <Button
            variant={statusFilter.includes("500") ? "secondary" : "outline"}
            onClick={() => handleStatusFilter("500")}
          >
            <AlertOctagonIcon className="mr-2 h-4 w-4" />
            500
          </Button>
          <Button
            variant={perceptionGroup ? "secondary" : "outline"}
            onClick={handleGroupBySimilar}
          >
            <GroupIcon className="mr-2 h-4 w-4" />
            Group by Similar
          </Button>
          <div className="flex items-center space-x-2 p-2">
            <Switch
              id="show-failed"
              checked={showFailed}
              onCheckedChange={handleToggleShowFailed}
            />
            <Label htmlFor="show-failed" className="text-sm">
              Show Failed
            </Label>
          </div>
        </div>
        <div className="flex items-center space-x-2">
          <Button
            variant="outline"
            size="icon"
            onClick={() => handlePageChange(page - 1)}
            disabled={page <= 1}
          >
            <ChevronLeftIcon className="h-4 w-4" />
          </Button>
          <Button
            variant="outline"
            size="icon"
            onClick={() => handlePageChange(page + 1)}
            disabled={page >= totalPages}
          >
            <ChevronRightIcon className="h-4 w-4" />
          </Button>
        </div>
      </div>

      <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4">
        {gallery?.map(screenshot => renderGalleryCard(screenshot))}
      </div>

      <div className="flex justify-between items-center mt-8">
        <Select value={limit.toString()} onValueChange={handleLimitChange}>
          <SelectTrigger className="w-[100px]">
            <SelectValue placeholder="Limit" />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="12">12</SelectItem>
            <SelectItem value="24">24</SelectItem>
            <SelectItem value="48">48</SelectItem>
            <SelectItem value="96">96</SelectItem>
          </SelectContent>
        </Select>

        <div className="flex items-center space-x-2">
          <Button
            variant="outline"
            size="sm"
            onClick={() => handlePageChange(1)}
            disabled={page <= 1}
          >
            First
          </Button>
          <Button
            variant="outline"
            size="sm"
            onClick={() => handlePageChange(page - 1)}
            disabled={page <= 1}
          >
            «
          </Button>
          {renderPageButtons(8)}
          <Button
            variant="outline"
            size="sm"
            onClick={() => handlePageChange(page + 1)}
            disabled={page >= totalPages}
          >
            »
          </Button>
          <Button
            variant="outline"
            size="sm"
            onClick={() => handlePageChange(totalPages)}
            disabled={page >= totalPages}
          >
            Last
          </Button>
        </div>
      </div>
    </div>
  );
};

export default GalleryPage;