import { blogs } from "../data/blog";
import Blog from "./Blog";

function BlogSection() {
  return (
    <div className="pb-11">
      <div className="flex flex-row items-baseline justify-between mb-4.5">
        <div className="text-[19px] mb-4.5 font-space font-semibold tracking-[-0.4px]">Writing</div>
        <div className="text-[13px] font-medium text-(--gold) cursor-pointer hover:text-(--gold-deep)">All Posts →</div>
      </div>
      <div>
        {blogs.map(blog => <Blog
          title = {blog.title}
          readTime= {blog.readTime}
          genere= {blog.genere}
          date= {blog.date}
        />)}
      </div>
    </div>
  );
}

export default BlogSection;
