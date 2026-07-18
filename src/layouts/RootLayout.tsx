import { Outlet } from "react-router-dom";
import Footer from "../components/Footer";
import Navbar from "../components/Navbar";

export function RootLayout() {
  return (
    <div className="min-h-screen flex flex-col bg-(--home-bg)">
      <Navbar />
      <div className="flex-1 flex flex-col max-w-(--maxw) mx-auto px-6 w-full">
        <Outlet />
      </div>
      <Footer />
    </div>
    )
}
