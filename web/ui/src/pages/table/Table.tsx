import { useEffect, useState, useMemo } from "react";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { WideSkeleton } from "@/components/loading";
import { Link } from "react-router-dom";
import { toast } from "@/hooks/use-toast";
import { Badge } from "@/components/ui/badge";
import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { ArrowUpDown, RefreshCw, XIcon } from "lucide-react";
import * as api from "@/lib/api/api";
import * as apitypes from "@/lib/api/types";

export default function TablePage() {
  const [loading, setLoading] = useState<boolean>(true);
  const [list, setList] = useState<apitypes.list[]>([]);
  const [searchTerm, setSearchTerm] = useState("");
  const [sortColumn, setSortColumn] = useState<keyof apitypes.list>("id");
  const [sortDirection, setSortDirection] = useState<"asc" | "desc">("asc");
  const [filterStatus, setFilterStatus] = useState<"all" | "success" | "error">("all");

  useEffect(() => {
    fetchData();
  }, []);

  const fetchData = async () => {
    setLoading(true);
    try {
      const s = await api.get('list');
      setList(s);
    } catch (err) {
      toast({
        title: "API Error",
        variant: "destructive",
        description: `Failed to get list: ${err}`
      });
    } finally {
      setLoading(false);
    }
  };

  const getStatusColor = (code: number) => {
    if (code >= 200 && code < 300) return "bg-green-500 text-white";
    if (code >= 400 && code < 500) return "bg-yellow-500 text-black";
    if (code >= 500) return "bg-red-500 text-white";
    return "bg-gray-500 text-white";
  };

  const handleSort = (column: keyof apitypes.list) => {
    if (column === sortColumn) {
      setSortDirection(sortDirection === "asc" ? "desc" : "asc");
    } else {
      setSortColumn(column);
      setSortDirection("asc");
    }
  };

  const copyToClipboard = (content: string, type: string) => {
    navigator.clipboard.writeText(content).then(() => {
      toast({
        description: `${type} copied to clipboard`,
      });
    }).catch((err) => {
      console.error('Failed to copy content: ', err);
      toast({
        title: "Error",
        description: "Failed to copy content",
        variant: "destructive",
      });
    });
  };

  const filteredAndSortedList = useMemo(() => {
    return list
      .filter((item) => {
        const matchesSearch =
          item.url.toLowerCase().includes(searchTerm.toLowerCase()) ||
          item.title.toLowerCase().includes(searchTerm.toLowerCase());
        const matchesStatus =
          filterStatus === "all" ||
          (filterStatus === "success" && item.response_code < 400) ||
          (filterStatus === "error" && item.response_code >= 400);
        return matchesSearch && matchesStatus;
      })
      .sort((a, b) => {
        if (a[sortColumn] < b[sortColumn]) return sortDirection === "asc" ? -1 : 1;
        if (a[sortColumn] > b[sortColumn]) return sortDirection === "asc" ? 1 : -1;
        return 0;
      });
  }, [list, searchTerm, sortColumn, sortDirection, filterStatus]);

  if (loading) return <WideSkeleton />;

  return (
    <>
      <div className="flex flex-col md:flex-row justify-between items-center space-y-2 md:space-y-0 md:space-x-2">
        <div className="flex items-center w-full md:w-auto">
          <Input
            placeholder="Filter by URL or title..."
            value={searchTerm}
            onChange={(e) => setSearchTerm(e.target.value)}
            className="mr-2"
          />
          <Button variant="outline" size="icon" onClick={() => setSearchTerm("")}>
            <XIcon className="h-4 w-4" />
          </Button>
        </div>
        <div className="flex items-center space-x-2 w-full md:w-auto">
          <Select value={filterStatus} onValueChange={(value: "all" | "success" | "error") => setFilterStatus(value)}>
            <SelectTrigger className="w-[180px]">
              <SelectValue placeholder="Filter by status" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="all">All</SelectItem>
              <SelectItem value="success">Success</SelectItem>
              <SelectItem value="error">Error</SelectItem>
            </SelectContent>
          </Select>
          <Button variant="outline" size="icon" onClick={fetchData}>
            <RefreshCw className="h-4 w-4" />
          </Button>
        </div>
      </div>
      <div className="rounded-md border">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead className="w-[100px]">Actions</TableHead>
              <TableHead className="w-[100px] cursor-pointer" onClick={() => handleSort("response_code")}>
                Code {sortColumn === "response_code" && <ArrowUpDown className="ml-2 h-4 w-4 inline" />}
              </TableHead>
              <TableHead className="cursor-pointer" onClick={() => handleSort("url")}>
                URL {sortColumn === "url" && <ArrowUpDown className="ml-2 h-4 w-4 inline" />}
              </TableHead>
              <TableHead className="cursor-pointer" onClick={() => handleSort("title")}>
                Title {sortColumn === "title" && <ArrowUpDown className="ml-2 h-4 w-4 inline" />}
              </TableHead>
              <TableHead className="cursor-pointer" onClick={() => handleSort("content_length")}>
                Size {sortColumn === "content_length" && <ArrowUpDown className="ml-2 h-4 w-4 inline" />}
              </TableHead>
              <TableHead>Protocol</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {filteredAndSortedList.map(item => (
              <TableRow key={item.id}>
                <TableCell>
                  <Link to={`/screenshot/${item.id}`} className="text-blue-600 hover:underline">
                    View
                  </Link>
                </TableCell>
                <TableCell>
                  <Badge variant="outline" className={`${getStatusColor(item.response_code)}`}>
                    {item.response_code}
                  </Badge>
                </TableCell>
                <TableCell
                  className="break-all cursor-pointer font-mono max-w-[300px] truncate"
                  onClick={() => copyToClipboard(item.url, 'URL')}
                >
                  {item.url}
                </TableCell>
                <TableCell
                  className="break-all cursor-pointer max-w-[300px] truncate"
                  onClick={() => copyToClipboard(item.title, 'Title')}
                >
                  {item.title}
                </TableCell>
                <TableCell>{(item.content_length / 1024).toFixed(2)} KB</TableCell>
                <TableCell>{item.protocol}</TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      </div>
      <div className="mt-4 text-sm text-gray-500">
        Showing {filteredAndSortedList.length} of {list.length} results
      </div>
    </>
  );
}