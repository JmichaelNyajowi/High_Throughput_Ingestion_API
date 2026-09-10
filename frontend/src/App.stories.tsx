import type { Meta, StoryObj } from "@storybook/react-vite";
import { App } from "./app/App";
import "./styles.css";

const meta = { title: "Operations/Application shell", component: App, parameters: { layout: "fullscreen" } } satisfies Meta<typeof App>;
export default meta;
type Story = StoryObj<typeof meta>;
export const FleetPending: Story = { args: { initialPath: "/" } };
export const DevicePending: Story = { args: { initialPath: "/devices/gateway-17" } };
export const HistoryPreQuery: Story = { args: { initialPath: "/devices/gateway-17/history" } };
export const NotFound: Story = { args: { initialPath: "/missing" } };
