import json
import os

current_folder = os.path.abspath("")
monster_path = current_folder + "\\actors.csv"
actors_path = current_folder+ "\\actors.json"

monsters = open(actors_path,"r",encoding="utf-8").read()

to_append = ""
with open(monster_path,"r", encoding="utf-8", ) as mon_book:
    for i,line in enumerate(mon_book.readlines()):
        i += 1
        stats = line[1:].strip().split(",")
        name = stats[0]
        leyenda = int(stats[len(stats)-1])
        stats = list(map(int, stats[1:]))
        newmob = f"""
        {{
            "name":"{name}",
            "level":0,
            "age":0,
            "legend_level":"0",
            "reach":"int",
            "picture":"Systems/role_up/media/images/placeholder.png",
            "main_class":"",
            "subclass":"",
            "description":"",
            "strengths":"",
            "weaknesess":"",
            "current_health":{stats[0]},
            "max_health":{stats[0]},
            "eva":{stats[1]},
            "imp":{stats[2]},
            "pun":{stats[3]},
            "mag":{stats[4]},
            "fza":{stats[5]},
            "res":{stats[6]},
            "agl":{stats[7]},
            "hab":{stats[8]},
            "per":{stats[9]},
            "int":{stats[10]},
            "vol":{stats[11]},
            "car":{stats[12]},
            "sue":{stats[13]},    
            "actions":[],
            "items":[]
        }},"""
        if(leyenda == -1):
            to_append += newmob

with open(actors_path,"a", encoding="utf-8", ) as mon_book:
    mon_book.write(to_append) 