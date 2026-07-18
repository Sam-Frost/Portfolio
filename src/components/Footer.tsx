import Button from "./Button";

function Footer() {
  return (
    <div className="px-6 py-9 text-center border-t-[0.5px] border-(--line-soft) bg-(--card)">
      <h2 className="mb-1.5 text-xl sapce-font font-semibold spacing tracking-[-0.4px]" >Let's build something together</h2>
      <p className="text-[13px] mb-4.5 text-(--navbar-link)">samarthnegi2002@gmail.com · usually replies within a day</p>
      <div className="flex flex-row gap-2.5 flex-wrap justify-center items-center ">
        <Button styling = "px-4.5 py-2"  text= "Get in touch"/>
        <Button styling = "px-4.5 py-2" colorTheme="bg-transparent  border-[0.5px] border-[#3A3936]"  text= "📅 Book a call"/>
      </div>
    </div>
  );
}

export default Footer;
