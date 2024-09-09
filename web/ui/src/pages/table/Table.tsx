import {
  useEffect,
  useState
} from "react";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow
} from "@/components/ui/table";
import { WideSkeleton } from "@/components/loading";
import { Link } from "react-router-dom";
import { toast } from "@/hooks/use-toast";
import { Badge } from "@/components/ui/badge";
import * as api from "@/lib/api";

const TablePage = () => {
  const [loading, setLoading] = useState<boolean>(true);
  const [list, setList] = useState<api.list[]>();

  useEffect(() => {
    const getData = async () => {
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
    getData();
  }, []);

  const getStatusColor = (code: number) => {
    if (code >= 200 && code < 300) return "bg-green-500 text-white";
    if (code >= 400 && code < 500) return "bg-yellow-500 text-black";
    if (code >= 500) return "bg-red-500 text-white";
    return "bg-gray-500 text-white";
  };

  if (loading) return <WideSkeleton />;
  if (!list) return;

  return (
    <Table>
      <TableHeader>
        <TableRow>
          <TableHead></TableHead>
          <TableHead>Code</TableHead>
          <TableHead>URL</TableHead>
          <TableHead>Title</TableHead>
        </TableRow>
      </TableHeader>
      <TableBody>
        {list.map(item => (
          <TableRow key={item.id}>
            <TableCell>
              <Link to={`/screenshot/${item.id}`} className="text-blue-600 hover:underline">
                View
              </Link>
            </TableCell>
            <TableCell>
              <Badge variant="default" className={`${getStatusColor(item.response_code)}`}>
                {item.response_code}
              </Badge>
            </TableCell>
            <TableCell className="font-mono">{item.url}</TableCell>
            <TableCell>{item.title}</TableCell>
          </TableRow>
        ))}
      </TableBody>
    </Table>
  );
};

export default TablePage;