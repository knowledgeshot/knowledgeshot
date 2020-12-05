# KnowledgeShot
##### This is where the API and Web side of things live!

The site is available on [knowledgeshot.de](https://knowledgeshot.de/).

On Version 0.6 - Feel free to contribute!

### Contribution
To add an article, see one of the JSONs in the pages/ submodule. Change the text there and make a Pull Request with the new article under the submodule. If it gets approved it will be visible soon!

To add features or fix bugs, check out the code and create a Pull Request!


### The API
The API has the following calls:
- /api/random
    - Displays the link and title to a random page
    - Can be used for a "Random Article" button as seen in the Web-Interface
- /api/search/{term}
    - {term} to be replaced with a URL-Save encoded string
    - Searches the sites database and gives back the answer
- /api/page/{term}
    - {term} to be replaced with the valid site id. Can be retrieved from /api/search or /api/random.
    - Gives back all data from a specific site to display content
    - You MUST include checks to see if all infos are filled as some articles, for example, don't have images or sources.
- /api/version
    - Displays the Backends Version for the program, the API and the minimum supported client version
    - When developing a program that uses the API, code the as of programming the latest version int. in and run checks if it is bigger than minimum supported.
        - This is to ensure your client doesn't break and to potentially warn users that stuff might not work.
    - In a perfect scenario, the current version and the client version int are equal.

The API has **no** key authentication. My goal is to make the API as open as possible. This might change in the future, depending on how many calls we get.

The API is available for all under `https://knowledgeshot.de/api/`.

### ToDos

- Add multi-image support (partially implemented)
- ~~Add author details (should be easy enough)~~
- ~~Redesign home screen with actual stats/info/help~~
- Redesign the article pages to look cleaner