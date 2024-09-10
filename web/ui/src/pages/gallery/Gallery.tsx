import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardFooter
} from "@/components/ui/card";
import {
  useEffect,
  useState
} from "react";
import {
  Link,
  useSearchParams
} from "react-router-dom";
import { toast } from "@/hooks/use-toast";
import { WideSkeleton } from "@/components/loading";
import { Badge } from "@/components/ui/badge";
import { ExternalLinkIcon } from "lucide-react";
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger
} from "@/components/ui/tooltip";
import * as api from "@/lib/api/api";
import * as apitypes from "@/lib/api/types";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue
} from "@/components/ui/select";
import { Input } from "@/components/ui/input";


const GalleryPage = () => {
  const [gallery, setGallery] = useState<apitypes.galleryResult[]>();
  const [wappalyzer, setWappalyzer] = useState<apitypes.wappalyzer>();
  const [totalPages, setTotalPages] = useState(0);
  const [loading, setLoading] = useState(true);

  const [searchParams, setSearchParams] = useSearchParams();
  const page = parseInt(searchParams.get("page") || "1");
  const limit = parseInt(searchParams.get("limit") || "24");

  useEffect(() => {
    const getData = async () => {
      try {
        const s = await api.get('wappalyzer');
        setWappalyzer(s);
      } catch (err) {
        toast({
          title: "API Error",
          variant: "destructive",
          description: `Failed to get wappalyzer: ${err}`
        });
      }
    };
    getData();
  }, []);

  useEffect(() => {
    const getData = async () => {
      setLoading(true);
      try {
        const s = await api.get('gallery', { page: page, limit: limit });
        setGallery(s.results);
        setTotalPages(Math.ceil(s.total_count / limit));
      } catch (err) {
        toast({
          title: "API Error",
          variant: "destructive",
          description: `Failed to get gallery: ${err}`
        });
      } finally {
        setLoading(false);
      }
    };
    getData();
  }, [page, limit]);

  const getStatusColor = (code: number) => {
    if (code >= 200 && code < 300) return "bg-green-500 text-white";
    if (code >= 400 && code < 500) return "bg-yellow-500 text-black";
    if (code >= 500) return "bg-red-500 text-white";
    return "bg-gray-500 text-white";
  };

  const getIconUrl = (tech: string): string | undefined => {
    if (!wappalyzer || !(tech in wappalyzer)) return undefined;

    return wappalyzer[tech];
  };

  const handlePageChange = (newPage: number) => {
    setSearchParams({ page: newPage.toString(), limit: limit.toString() });
  };

  const handleLimitChange = (newLimit: string) => {
    setSearchParams({ limit: newLimit });
  };

  const renderPageButtons = () => {
    const pageButtons = [];
    const maxVisiblePages = 5; // Number of visible pages around the current page
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

  if (loading) return <WideSkeleton />;

  return (
    <div className="space-y-6">

      <div className="flex flex-wrap gap-4 items-center justify-between rounded-lg">
        <div className="flex flex-wrap gap-2">
          <Select >
            <SelectTrigger className="w-[200px]">
              <SelectValue placeholder="Technology Filter" />
            </SelectTrigger>
            <SelectContent>
              {/* {technologies.map((tech) => (
              <SelectItem key={tech} value={tech}>
                {tech}
              </SelectItem>
            ))} */}
            </SelectContent>
          </Select>
          <Button variant="outline">
            HTTP 200
          </Button>
          <Button variant="secondary">
            HTTP 400
          </Button>
          <div className="flex gap-2">
            <Input
              type="number"
              placeholder="Custom Code"
              value="custom status"
              className="w-32"
            />
            <Button >Filter</Button>
          </div>
        </div>
        <Button variant="outline">
          Group by Similar
        </Button>
      </div>

      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4">
        {gallery?.map(screenshot => (
          <Link to={`/screenshot/${screenshot.id}`} key={screenshot.id}>
            <Card className="group overflow-hidden transition-all hover:shadow-lg">
              <CardContent className="p-0 relative">
                <img
                  src={api.endpoints.screenshot.path + "/" + screenshot.file_name}
                  alt={screenshot.url}
                  loading="lazy"
                  className="w-full h-48 object-cover transition-all duration-300 filter group-hover:scale-105"
                />
                <div className="absolute top-2 right-2">
                  <Badge variant="default" className={`${getStatusColor(screenshot.response_code)}`}>
                    {screenshot.response_code}
                  </Badge>
                </div>
                <div className="absolute bottom-2 right-2 opacity-0 group-hover:opacity-100 transition-opacity">
                  <ExternalLinkIcon className="text-white drop-shadow-lg" />
                </div>
              </CardContent>

              <CardFooter className="p-2 flex items-start justify-between">
                <div className="flex-grow mr-2">
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
                <div className="flex flex-shrink-0 space-x-1">
                  {screenshot.technologies?.slice(0, 5).map(tech => {
                    const iconUrl = getIconUrl(tech);
                    return iconUrl ? (
                      <TooltipProvider key={tech}>
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
              </CardFooter>
            </Card>
          </Link>
        ))}
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
          {renderPageButtons()}
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