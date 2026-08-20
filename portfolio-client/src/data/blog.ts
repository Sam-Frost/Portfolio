import imageBlurring from "../assets/blogs/image-bluring-at-client-vs-server.md?raw"

type Blog = {
  title: string
  fileName: string
  readTime: string
  genere: string
  date: string
  content: string
}

export const blogs: Blog[] = [{
  title: "Image bluring at client vs server",
  fileName: "image-bluring-at-client-vs-server",
  readTime: "5 min read",
  genere: "Backend",
  date: "Dec 2024",
  content: imageBlurring
},
]


export function getBlogData(fileName: string): string {
  const blog = blogs.find(blog => blog.fileName == fileName)

  if (!blog)
    throw new Error(`Blog file does not exist for ${fileName}`)

  return blog.content;
}
