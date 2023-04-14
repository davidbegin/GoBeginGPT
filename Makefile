default:
	go build -o ./bin/$(NAME) . 

# These are all the folders and files we need to get started
# TODO: finish updating these
init:
	mkdir -p tmp
	mkdir -p skybox_archive
