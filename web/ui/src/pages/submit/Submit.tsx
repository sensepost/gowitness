import { useEffect, useState } from 'react';
import { Form, useActionData, useNavigation } from 'react-router-dom';
import { PlusCircle, Trash2, Send, Settings, GlobeIcon, ExternalLinkIcon, ServerIcon, FileTypeIcon, ClockIcon } from 'lucide-react';
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Switch } from "@/components/ui/switch";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Accordion, AccordionContent, AccordionItem, AccordionTrigger } from "@/components/ui/accordion";
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle } from '@/components/ui/dialog';
import { ScrollArea } from '@/components/ui/scroll-area';
import * as apitypes from "@/lib/api/types";

export default function JobSubmissionPage() {
  const [urls, setUrls] = useState<string[]>(['']);
  const [advancedOptions, setAdvancedOptions] = useState(false);
  const [immediateUrl, setImmediateUrl] = useState<string>('');
  const [isModalOpen, setIsModalOpen] = useState<boolean>(false);

  const navigation = useNavigation();
  const probeResult = useActionData() as apitypes.detail | null;

  useEffect(() => {
    probeResult
      ? setIsModalOpen(true)
      : setIsModalOpen(false);
  }, [probeResult]);

  const handleUrlChange = (index: number, value: string) => {
    const newUrls = [...urls];
    newUrls[index] = value;
    setUrls(newUrls);
  };

  const addUrl = () => {
    setUrls([...urls, '']);
  };

  const removeUrl = (index: number) => {
    const newUrls = urls.filter((_, i) => i !== index);
    setUrls(newUrls);
  };

  const ProbeOptions = () => (
    <Accordion type="single" collapsible>
      <AccordionItem value="options">
        <AccordionTrigger>
          <div className="flex items-center">
            <Settings className="mr-2 h-4 w-4" />
            Probe Options
          </div>
        </AccordionTrigger>
        <AccordionContent>
          <div className="space-y-4 pt-4">
            <div className="grid gap-4 md:grid-cols-3">
              <div className="space-y-2">
                <Label htmlFor="format">Screenshot Format</Label>
                <Select name="format" defaultValue="jpeg">
                  <SelectTrigger id="format">
                    <SelectValue placeholder="Select format" />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="png">PNG</SelectItem>
                    <SelectItem value="jpeg">JPEG</SelectItem>
                    <SelectItem value="pdf">PDF</SelectItem>
                  </SelectContent>
                </Select>
              </div>

              <div className="space-y-2">
                <Label htmlFor="timeout">Timeout (seconds)</Label>
                <Input
                  id="timeout"
                  name="timeout"
                  type="number"
                  min="0"
                  defaultValue="60"
                />
              </div>
              <div className="space-y-2">
                <Label htmlFor="delay">Screenshot Delay (seconds)</Label>
                <Input
                  id="delay"
                  name="delay"
                  type="number"
                  min="0"
                  defaultValue="5"
                />
              </div>
            </div>

            <div className="flex items-center space-x-2">
              <Switch
                id="advanced-options"
                checked={advancedOptions}
                onCheckedChange={setAdvancedOptions}
              />
              <Label htmlFor="advanced-options">Advanced Options</Label>
            </div>

            {advancedOptions && (
              <div className="space-y-4">
                <div className="space-y-2">
                  <Label htmlFor="user-agent">User Agent</Label>
                  <Input
                    id="user-agent"
                    name="user_agent"
                    defaultValue="Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/128.0.0.0 Safari/537.36"
                  />
                </div>

                <div className="grid gap-4 sm:grid-cols-2">
                  <div className="space-y-2">
                    <Label htmlFor="window-x">Window Width</Label>
                    <Input
                      id="window-x"
                      name="window_x"
                      type="number"
                      min="0"
                      defaultValue="1920"
                    />
                  </div>
                  <div className="space-y-2">
                    <Label htmlFor="window-y">Window Height</Label>
                    <Input
                      id="window-y"
                      name="window_y"
                      type="number"
                      min="0"
                      defaultValue="1080"
                    />
                  </div>
                </div>
              </div>
            )}
          </div>
        </AccordionContent>
      </AccordionItem>
    </Accordion>
  );

  return (
    <div className="container mx-auto py-6">
      <Card>
        <CardHeader>
          <CardTitle>Launch a New Probe</CardTitle>
          <CardDescription>Submit a job or run an immediate probe</CardDescription>
        </CardHeader>
        <CardContent>

          <Tabs defaultValue="job">
            <TabsList className="grid w-full grid-cols-2">
              <TabsTrigger value="job">Job Submission Probe</TabsTrigger>
              <TabsTrigger value="immediate">Immediate Probe</TabsTrigger>
            </TabsList>
            <TabsContent value="job">
              <Form method="post" className="space-y-6">
                <input type="hidden" name="action" value="job" />
                <div className="space-y-4">
                  <h3 className="text-lg font-semibold">URLs</h3>
                  {urls.map((url, index) => (
                    <div key={index} className="flex items-center space-x-2">
                      <Input
                        type="url"
                        name={`url-${index}`}
                        placeholder="https://sensepost.com"
                        value={url}
                        onChange={(e) => handleUrlChange(index, e.target.value)}
                        className="flex-grow"
                      />
                      {index === urls.length - 1 ? (
                        <Button type="button" variant="outline" size="icon" onClick={addUrl}>
                          <PlusCircle className="h-4 w-4" />
                        </Button>
                      ) : (
                        <Button type="button" variant="outline" size="icon" onClick={() => removeUrl(index)}>
                          <Trash2 className="h-4 w-4" />
                        </Button>
                      )}
                    </div>
                  ))}
                </div>

                <ProbeOptions />

                {!advancedOptions && (
                  <>
                    <input type="hidden" name="format" value="jpeg" />
                    <input type="hidden" name="timeout" value="60" />
                    <input type="hidden" name="delay" value="5" />
                    <input type="hidden" name="user_agent" value="Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/128.0.0.0 Safari/537.36" />
                    <input type="hidden" name="window_x" value="1920" />
                    <input type="hidden" name="window_y" value="1080" />
                  </>
                )}

                <input type="hidden" name="action" value="job" />

                <div className="flex justify-end">
                  <Button type="submit" disabled={navigation.state === "submitting"}>
                    <Send className="mr-2 h-4 w-4" />
                    {navigation.state === "submitting" ? "Submitting..." : `Submit ${urls.length} Target${urls.length > 1 ? "s" : ""}`}
                  </Button>
                </div>
              </Form>
            </TabsContent>
            <TabsContent value="immediate">
              <Form method="post" className="space-y-6">
                <div className="space-y-4">
                  <h3 className="text-lg font-semibold">URL</h3>
                  <Input
                    type="url"
                    name="immediate-url"
                    placeholder="https://sensepost.com"
                    value={immediateUrl}
                    onChange={(e) => setImmediateUrl(e.target.value)}
                    className="flex-grow"
                  />
                </div>

                <ProbeOptions />

                {!advancedOptions && (
                  <>
                    <input type="hidden" name="format" value="jpeg" />
                    <input type="hidden" name="timeout" value="60" />
                    <input type="hidden" name="delay" value="5" />
                    <input type="hidden" name="user_agent" value="Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/128.0.0.0 Safari/537.36" />
                    <input type="hidden" name="window_x" value="1920" />
                    <input type="hidden" name="window_y" value="1080" />
                  </>
                )}

                <input type="hidden" name="action" value="immediate" />

                <div className="flex justify-end">
                  <Button type="submit" disabled={navigation.state === "submitting"}>
                    <Send className="mr-2 h-4 w-4" />
                    {navigation.state === "submitting" ? "Running Probe..." : "Run Immediate Probe"}
                  </Button>
                </div>
              </Form>
            </TabsContent>
          </Tabs>
        </CardContent>
      </Card>

      <Dialog open={isModalOpen} onOpenChange={setIsModalOpen}>
        <DialogContent className="max-w-7xl max-h-[90vh] overflow-hidden flex flex-col">
          <DialogHeader>
            <DialogTitle>Probe Result</DialogTitle>
            <DialogDescription>Details of the immediate probe</DialogDescription>
          </DialogHeader>
          {probeResult && (
            <div className="flex-1 overflow-hidden">
              <div className="grid grid-cols-1 md:grid-cols-2 gap-6 h-full">
                <ScrollArea className="h-[calc(90vh-8rem)] pr-4">
                  <div className="space-y-6">
                    <Card>
                      <CardHeader>
                        <CardTitle className="text-lg font-semibold flex items-center">
                          <GlobeIcon className="mr-2 h-5 w-5" />
                          URL Information
                        </CardTitle>
                      </CardHeader>
                      <CardContent className="space-y-2">
                        <div className="flex justify-between">
                          <span className="font-medium">Initial URL:</span>
                          <a href={probeResult.url} target="_blank" rel="noopener noreferrer" className="text-blue-500 hover:underline flex items-center">
                            {probeResult.url}
                            <ExternalLinkIcon className="ml-1 h-3 w-3" />
                          </a>
                        </div>
                      </CardContent>
                    </Card>

                    <Card>
                      <CardHeader>
                        <CardTitle className="text-lg font-semibold flex items-center">
                          <ServerIcon className="mr-2 h-5 w-5" />
                          Response Details
                        </CardTitle>
                      </CardHeader>
                      <CardContent className="space-y-2">
                        <div className="flex justify-between">
                          <span className="font-medium">Response Code:</span>
                          <span>{probeResult.response_code}</span>
                        </div>
                        <div className="flex justify-between">
                          <span className="font-medium">Protocol:</span>
                          <span>{probeResult.protocol}</span>
                        </div>
                        <div className="flex justify-between">
                          <span className="font-medium">Content Length:</span>
                          <span>{probeResult.content_length} bytes</span>
                        </div>
                      </CardContent>
                    </Card>

                    <Card>
                      <CardHeader>
                        <CardTitle className="text-lg font-semibold flex items-center">
                          <FileTypeIcon className="mr-2 h-5 w-5" />
                          Page Information
                        </CardTitle>
                      </CardHeader>
                      <CardContent className="space-y-2">
                        <div className="flex justify-between">
                          <span className="font-medium">Title:</span>
                          <span>{probeResult.title}</span>
                        </div>
                        <div className="flex justify-between">
                          <span className="font-medium">Failed:</span>
                          <span>{probeResult.failed ? 'Yes' : 'No'}</span>
                        </div>
                        {probeResult.failed && (
                          <div className="flex justify-between">
                            <span className="font-medium">Failed Reason:</span>
                            <span>{probeResult.failed_reason}</span>
                          </div>
                        )}
                      </CardContent>
                    </Card>

                    <Card>
                      <CardHeader>
                        <CardTitle className="text-lg font-semibold flex items-center">
                          <ClockIcon className="mr-2 h-5 w-5" />
                          Timing Information
                        </CardTitle>
                      </CardHeader>
                      <CardContent>
                        <div className="flex justify-between">
                          <span className="font-medium">Probed At:</span>
                          <span>{new Date(probeResult.probed_at).toLocaleString()}</span>
                        </div>
                      </CardContent>
                    </Card>
                  </div>
                </ScrollArea>
                <div className="h-[calc(90vh-8rem)] flex flex-col">
                  <h3 className="text-lg font-semibold mb-2">Screenshot</h3>
                  <div className="flex-1 overflow-hidden">
                    <img
                      src={`data:image/jpeg;base64,${probeResult.screenshot}`}
                      alt="Screenshot"
                      className="w-full h-full object-contain"
                    />
                  </div>
                </div>
              </div>
            </div>
          )}
        </DialogContent>
      </Dialog>

    </div>
  );
}