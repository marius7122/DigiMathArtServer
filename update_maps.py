import glob
from subprocess import call

maps_folder = 'files/maps'
backup_folder = 'files/maps_backup'

def copy_dir(src, dest):
	call(['cp', '-r', src, dest])

# make backup folder 
copy_dir(maps_folder, backup_folder)


def update_map(map):
	return map.replace('"shape"', '"shapeIndex"')



maps_path = glob.glob(maps_folder + '/*.json')
for map_path in maps_path:
	file = open(map_path, 'r')
	updated_map = update_map(file.read())
	file.close()

	file = open(map_path, 'w')
	file.write(updated_map)
	file.close()
