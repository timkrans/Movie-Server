# Movie Server
- This prorject currently is a backend only movie server

## Curent Endpoints
```
POST /movies 
```
- Allows for creation of movies in the sql lite db with meta data and location of the movie which must already exist on disk
- Params for this and the put one are
```title , video_file_path, cover_image_file_path```
and the cover is not required

```
GET /movies/:id
```
- Return the byte file to actually play the video 
```
PUT /movies/:id   
```
- Allows you to change the sqllite meta data for movie file location, name, cover location.. etc
```
DELETE /movies/:id 
```
- Will delete the meta data for the sql lite so movie is no longer able to be accessed by api

## Testing
- unit testing coming soon but can run a health check ```go test -v ./testing```

## Coming Soon
- Save state of video so I want to add params to start the video at a certain time
- Network connection so autoscalling the video output