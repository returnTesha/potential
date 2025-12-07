import Image from "next/image"

import {
  SidebarInset,
  SidebarProvider,
} from "@/registry/new-york-v4/ui/sidebar"
import { AppSidebar } from "@/app/(app)/main/dashboard/components/app-sidebar"
import { ChartAreaInteractive } from "@/app/(app)/main/dashboard/components/chart-area-interactive"
import { DataTable } from "@/app/(app)/main/dashboard/components/data-table"
import { SectionCards }  from "@/app/(app)/main/dashboard/components/section-cards"
import { SiteHeader } from "@/app/(app)/main/dashboard/components/site-header"

import data from "./data.json"
import {ExamplesNav} from "@/components/examples-nav";
import {ThemeSelector} from "@/components/theme-selector";
import {PageNav} from "@/components/page-nav";
import EdgesFlow  from "@/app/(app)/main/dashboard/components/animated_test"

export default function Page() {
  return (
    <>
      <div className="md:hidden">
        <Image
          src="/examples/dashboard-light.png"
          width={1280}
          height={843}
          alt="Authentication"
          className="block dark:hidden"
          priority
        />
        <Image
          src="/examples/dashboard-dark.png"
          width={1280}
          height={843}
          alt="Authentication"
          className="hidden dark:block"
          priority
        />
      </div>
      <SidebarProvider
        className="hidden md:flex"
        style={
          {
            "--sidebar-width": "calc(var(--spacing) * 64)",
            "--header-height": "calc(var(--spacing) * 12 + 1px)",
          } as React.CSSProperties
        }
      >
        <AppSidebar variant="sidebar" />
        <SidebarInset>
          <SiteHeader />
          <div className="flex flex-1 flex-col">
            <div className="@container/main flex flex-1 flex-col gap-2">
              <div className="flex flex-col gap-4 py-4 md:gap-6 md:py-6">
                {/*<SectionCards />*/}
                {/*TODO react flow make here it.*/}
                <EdgesFlow />
                <div className="px-4 lg:px-6">
                  <ChartAreaInteractive />
                </div>
                <DataTable data={data} />
              </div>
            </div>
          </div>
        </SidebarInset>
      </SidebarProvider>
    </>
  )
}
