import { useContent } from "../content/ContentContext";
import { visibleSorted } from "../content/types";
import Blog from "./Blog";

function BlogSection() {
  const blogs = visibleSorted(useContent().blogs);
  return (
    <div className="pb-11">
      <div className="flex flex-row items-baseline justify-between mb-4.5">
        <div className="text-xl mb-4.5 font-space font-semibold tracking-[-0.4px]">Writing</div>
        <div className="text-[length:var(--text-caption)] font-medium text-(--gold) cursor-pointer hover:text-(--gold-deep)">All Posts →</div>
      </div>
      <div>
        {blogs.map((blog) => (
          <Blog
            key={blog.id}
            title={blog.title}
            slug={blog.slug}
            readTime={blog.readTime}
            genre={blog.genre}
            date={blog.date}
          />
        ))}
      </div>
    </div>
  );
}

export default BlogSection;
