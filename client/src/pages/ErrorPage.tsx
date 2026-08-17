import errorImage from "../assets/images/error.png"

import { useNavigate } from "react-router-dom"
import Navbar from "../components/Navbar"
import Footer from "../components/Footer"

export function ErrorPage() {

  const navigate = useNavigate()

  return (
    <div className="min-h-screen flex flex-col bg-(--bg)">
      <Navbar />
      <div className="flex-1 flex flex-col max-w-(--maxw) mx-auto px-6 w-full">
        <div className="flex-1 flex flex-col items-center justify-center">
        <img className="h-160 w-160 rounded-lg" src={errorImage} />
        <div className="mt-10 text-lg text-(--text-muted) cursor-pointer hover:text-(--fg)" onClick={() => {
          navigate("/")
        }}>Let's go to home... ?</div>
      </div>
      </div>
      <Footer />
    </div>

   )
}
