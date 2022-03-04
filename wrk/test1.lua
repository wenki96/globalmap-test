-- example script that demonstrates use of setup() to pass
-- data to and from the threads

local counter = 1
local threads = {}

function setup(thread)
-- 给每个线程设置一个 id 参数
   thread:set("id", counter)
-- 将线程添加到 table 中
   table.insert(threads, thread)
   counter = counter + 1
end

function init(args)
-- 初始化两个参数，每个线程都有独立的 requests、responses 参数
   requests  = 0
   responses = 0

-- 打印线程被创建的消息，打印完后，线程正式启动运行
   local msg = "thread %d created"
   print(msg:format(id))
end

function request()
-- 每发起一次请求 +1
   requests = requests + 1
   return wrk.request()
end

function response(status, headers, body)
-- 每得到一次请求的响应 +1
   if status == 200 then
      responses = responses + 1
   end
end

function done(summary, latency, requests)
-- 循环线程 table
   for index, thread in ipairs(threads) do
      local id        = thread:get("id")
      local requests  = thread:get("requests")
      local responses = thread:get("responses")
      local msg = "thread %d made %d requests and got %d responses"
-- 打印每个线程发起了多少个请求，得到了多少次响应
      print(msg:format(id, requests, responses))
   end
end