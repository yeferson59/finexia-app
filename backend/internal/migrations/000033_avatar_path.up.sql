-- Store the avatar as a path, not as an absolute URL.
--
-- UploadAvatarToS3 used to write PUBLIC_URL + "/users/<id>/avatar", which put
-- the API's own host into every row: https://api.finexia.me in production. The
-- API now lives on a private network and the only door to it from outside is
-- the web app, which serves the avatar at that same path
-- (frontend/src/lib/api/proxy.ts, docs/API.md §1.6). A row that names the old
-- host names a server nobody can reach, and the browser shows a broken image.
--
-- The path is what the application writes from now on, and unlike a host it
-- survives the next move too. Only the URL the upload built for that same user
-- is rewritten: an image a user pointed somewhere else by hand is left alone.
UPDATE users
SET image = '/users/' || id::text || '/avatar'
WHERE image ~ ('^https?://[^/]+/users/' || id::text || '/avatar$');
