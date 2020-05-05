########################################################################
# File Name: merge_file_blocks.sh
# Author: fangjiamou
# mail: fangjiamou@gmail.com
# Created Time: 2020年05月05日 星期二 15时37分10秒
########################################################################
#!/bin/bash

#　获取传入脚本的参数
filepath=$1
filestore=$2

# echo "filepath: " $filepath
# echo "filestore: " $filestore

if [ ! -f $filestore ]; then
    echo "$filestore not exist"
else
    rm -f $filestore
fi


#for item in `ls $filepath | sort -n`
#do
#    `cat $filepath/${item} >> ${filestore}
#    echo "merge ${filepath}/${item} to $filestore ok"
#done

# echo "file merge ok"




