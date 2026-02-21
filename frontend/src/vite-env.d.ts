/// <reference types="vite/client" />

declare module '*.vue' {
    import type {DefineComponent} from 'vue'
    const component: DefineComponent<{}, {}, any>
    export default component
}

// Wails runtime type declaration
interface WailsRuntime {
    LogPrint(message: string): void;
    LogTrace(message: string): void;
    LogDebug(message: string): void;
    LogInfo(message: string): void;
    LogWarning(message: string): void;
    LogError(message: string): void;
    LogFatal(message: string): void;
    EventsOnMultiple(eventName: string, callback: (...data: any) => void, maxCallbacks: number): () => void;
    EventsOff(eventName: string, ...additionalEventNames: string[]): void;
    EventsOffAll(): void;
    EventsEmit(eventName: string, ...data: any): void;
    WindowReload(): void;
    WindowReloadApp(): void;
    WindowSetAlwaysOnTop(b: boolean): void;
    WindowSetSystemDefaultTheme(): void;
    WindowSetLightTheme(): void;
    WindowSetDarkTheme(): void;
    WindowCenter(): void;
    WindowSetTitle(title: string): void;
    WindowFullscreen(): void;
    WindowUnfullscreen(): void;
    WindowIsFullscreen(): Promise<boolean>;
    WindowSetSize(width: number, height: number): void;
    WindowGetSize(): Promise<{ w: number; h: number }>;
    WindowSetMaxSize(width: number, height: number): void;
    WindowSetMinSize(width: number, height: number): void;
    WindowSetPosition(x: number, y: number): void;
    WindowGetPosition(): Promise<{ x: number; y: number }>;
    WindowHide(): void;
    WindowShow(): void;
    WindowMaximise(): void;
    WindowToggleMaximise(): void;
    WindowUnmaximise(): void;
    WindowIsMaximised(): Promise<boolean>;
    WindowMinimise(): void;
    WindowUnminimise(): void;
    WindowIsMinimised(): Promise<boolean>;
    WindowIsNormal(): Promise<boolean>;
    WindowSetBackgroundColour(R: number, G: number, B: number, A: number): void;
    ScreenGetAll(): Promise<{ isCurrent: boolean; isPrimary: boolean; width: number; height: number }[]>;
    BrowserOpenURL(url: string): void;
    Environment(): Promise<{ buildType: string; platform: string; arch: string }>;
    Quit(): void;
    Hide(): void;
    Show(): void;
    ClipboardGetText(): Promise<string>;
    ClipboardSetText(text: string): Promise<boolean>;
    OnFileDrop(callback: (x: number, y: number, paths: string[]) => void, useDropTarget?: boolean): void;
    OnFileDropOff(): void;
    CanResolveFilePaths(): boolean;
    ResolveFilePaths(files: File[]): void;
}

declare global {
    interface Window {
        runtime: WailsRuntime;
    }
}

export {}
