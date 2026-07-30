I was building a matrimonial platform when I came across this issue of image blurring. We needed to make sure the images of the user’s are blurred before they hit the client in order to maintain privacy of the user.

I encountered this again while browsing LinkedIn. I noticed that the “Viewers you might be interested in” section had images blurred from the backend, likely because it’s a paid feature. If users could access the data just by inspecting elements, it wouldn’t be much of a paywall. In contrast, the “People you may know” section applied the blur effect on the client side, making it easy to reveal image URLs by tinkering with the inspect element.

This experience led me to explore the differences between client-side and server-side blurring.

Client Side Blurring
Client-side blurring is the process of blurring the image on the client’s machine using CSS or javascript running on the browser.

The key advantage of client-side blurring is offloading the processing from the server. While blurring may seem lightweight initially, at scale, handling thousands of simultaneous requests can create significant computational overhead.

Server Side Blurring
Server-side blurring involves processing the image on the server before sending it to the client.

Press enter or click to view image in full size

Blurring image before sending to client
Another approach is to generate and store two versions of the image at the time of upload, allowing the system to serve the appropriate version as needed.

Write on Medium
This approach is more efficient than the previous one since the image is blurred only once and stored, allowing us to serve the pre-processed version directly to the client. This reduces computational overhead and minimize latency whenever the client requests the image.

Press enter or click to view image in full size

Blurring image before saving to S3
Why should one prefer server side blurring :

Client-side blurring is limited to CSS and JavaScript, while server-side processing allows the use of custom libraries and different programming languages.
Server-side blurring offers more control over image modifications, enabling advanced processing beyond just blurring.
Conclusion
Image blurring can be a computationally expensive task. It is preferred to offload this computationally expensive task to the client whenever possible.

Of course not everything can be offload to the client. If there is some sensitive information(for eg : people who have viewed your profile on linkedIn), which shall not be accessible to the user or is hidden behind a pay wall.

Press enter or click to view image in full size

Profile photos are blurred and no hyperlinks present to the user profiles
So general approach would be offload blurring tasks as much as we can to the client, but whenever we have to deal with sensitive information the server’s have to bear extra computational cost of blurring.
