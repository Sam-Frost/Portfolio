import { useNavigate } from "react-router-dom";
import badUrlImage from "../assets/images/bad-url.png"

export function BadUrlPage() {

  const navigate = useNavigate();

  return (
    <div className="flex-1 flex flex-col items-center justify-center">
      <img className="h-160 w-160 rounded-lg" src={badUrlImage} />
      <div className="mt-10 text-lg text-(--navbar-link) cursor-pointer hover:text-(--button)" onClick={() => {
        navigate("/")
      }}>Let's go to home... ?</div>
    </div>
  )
}
