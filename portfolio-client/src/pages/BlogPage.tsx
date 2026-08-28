import { useEffect } from "react"
import { useNavigate, useParams } from "react-router-dom";
import ReactMarkdown from "react-markdown";
import remarkGfm from "remark-gfm";
import { useContent, useContentLoaded } from "../content/ContentContext";

export function BlogPage() {
  const { blogSlug } = useParams();
  const navigate = useNavigate();
  const loaded = useContentLoaded();
  const blogs = useContent().blogs;
  const blog = blogs.find((b) => b.slug === blogSlug && b.visible);

  useEffect(() => {
    if (loaded && !blog) navigate("/invalid-url");
  }, [loaded, blog, navigate]);

  return (
    <article className="prose prose-invert mx-auto max-w-(--maxw) py-10">
      <ReactMarkdown remarkPlugins={[remarkGfm]}>{blog?.body ?? ""}</ReactMarkdown>
    </article>
  )
}
