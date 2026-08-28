import { useNavigate } from "react-router-dom"

interface Blog {
  title: string
  slug: string
  readTime: string
  genre: string
  date: string
}

function Blog(props: Blog) {

  const navigate = useNavigate();
  return (
    <div className="hover:text-(--gold) cursor-pointer flex justify-between items-center py-3.5 gap-3 border-b-[0.5px] border-solid border-(--line-soft)" onClick={() => {
      navigate(`/blogs/${props.slug}`)
    }}>
      <div>
        <div className="font-medium text-sm">{props.title}</div>
        <div className="text-xs mt-1 text-(--text-muted)">{props.readTime} · {props.genre}</div>
      </div>
      <div className="text-xs text-(--text-muted)">{props.date}</div>
    </div>
  );
}

export default Blog;
