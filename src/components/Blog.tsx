interface Blog {
  title: string
  readTime: string
  genere: string
  date: string
}

function Blog(props: Blog) {
  return (
    <div className="hover:text-(--gold) cursor-pointer flex justify-between items-center py-3.5 gap-3 border-b-[0.5px] border-solid border-(--line-soft)">
      <div>
        <div className="font-medium text-sm">{props.title}</div>
        <div className="text-xs mt-1 text-(--navbar-link)">{props.readTime} · {props.genere}</div>
      </div>
      <div className="text-xs text-(--navbar-link)">{props.date}</div>
    </div>
  );
}

export default Blog;
