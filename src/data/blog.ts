import testBlog from "../assets/blogs/test-blog.md?raw"

type Blog = {
  title: string
  readTime: string
  genere: string
  date: string
  content: string
}

export const blogs: Blog[] = [{
  title : "Test Blog but with a long TITLE.....",
  readTime: "8 min read",
  genere: "Databases",
  date: "Jun 2026",
  content : testBlog
},
{
  title : "Test Blog but with a long TITLE.....",
  readTime: "8 min read",
  genere: "Databases",
  date: "Jun 2026",
  content : testBlog
  },
  {
    title : "Test Blog but with a long TITLE.....",
    readTime: "8 min read",
    genere: "Databases",
    date: "Jun 2026",
    content : testBlog
  }]
