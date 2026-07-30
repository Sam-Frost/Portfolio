import { useEffect, useState } from "react"
import { useLocation, useNavigate } from "react-router-dom";
import ReactMarkdown from "react-markdown";
import remarkGfm from "remark-gfm";
import { getBlogData } from "../data/blog";

export function BlogPage() {

  const location = useLocation();
  const [data, setData] = useState("")
  const navigate = useNavigate();

  useEffect(() => {
    try {
      const fileName = location.pathname.substring(7) // Remove "/blogs" from url
      const blogData = getBlogData(fileName)
      setData(blogData)
    } catch (e) {
      navigate("/invalid-url")
    }
  })

  return (
    <article className="prose prose-invert mx-auto max-w-(--maxw) py-10">
      <ReactMarkdown remarkPlugins={[remarkGfm]}>{data}</ReactMarkdown>
    </article>
  )
}
