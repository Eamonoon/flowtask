import Link from "next/link";

export default function Home() {
  return (
    <main className="flex flex-col flex-1 items-center justify-center min-h-screen">
      <div className="text-center space-y-6">
        <h1 className="text-4xl font-bold">FlowTask</h1>
        <p className="text-xl text-muted-foreground">
          AI 驱动的学习计划和任务管理平台
        </p>
        <div className="flex gap-4 justify-center">
          <Link
            href="/register"
            className="px-6 py-3 bg-primary text-primary-foreground rounded-md hover:opacity-90"
          >
            开始使用
          </Link>
          <Link
            href="/login"
            className="px-6 py-3 border border-primary text-primary rounded-md hover:bg-accent"
          >
            登录
          </Link>
        </div>
      </div>
    </main>
  );
}
