import type { Metadata } from "next";
import "./globals.css";

export const metadata: Metadata = {
  title: "QueryBase - Database Explorer",
  description: "Query and manage your databases with approval workflow",
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="en">
      <body className="antialiased">
        {children}
      </body>
    </html>
  );
}
