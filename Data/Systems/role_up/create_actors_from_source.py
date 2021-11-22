import json
import os

current_folder = os.path.abspath("")
monster_path = current_folder + "\\monsters.csv"
actors_path = current_folder+ "\\actors.json"

monsters = open(actors_path,"r",encoding="utf-8").read()

to_append = ""
with open(monster_path,"r", encoding="utf-8", ) as mon_book:
    for line in mon_book.readlines():
        stats = line[1:].strip().split(",")
        name = stats[0]
        leyenda = int(stats[len(stats)-1])
        stats = list(map(int, stats[1:]))
        newmob = f"""
        {{
            "name":"{name}",
            "description":"",
            "strengths":"",
            "weaknesess":"",
            "extra":"",
            "stats":[
                {{
                    "name":"ps",
                    "ammount":{stats[0]}
                }},
                {{
                    "name":"eva",
                    "ammount":{stats[1]}
                }},
                {{
                    "name":"imp",
                    "ammount":{stats[2]}
                }},
                {{
                    "name":"pun",
                    "ammount":{stats[3]}
                }},
                {{
                    "name":"mag",
                    "ammount":{stats[4]}
                }},
                {{
                    "name":"fza",
                    "ammount":{stats[5]}
                }},
                {{
                    "name":"res",
                    "ammount":{stats[6]}
                }},
                {{
                    "name":"agl",
                    "ammount":{stats[7]}
                }},
                {{
                    "name":"hab",
                    "ammount":{stats[8]}
                }},
                {{
                    "name":"per",
                    "ammount":{stats[9]}
                }},
                {{
                    "name":"int",
                    "ammount":{stats[10]}
                }},
                {{
                    "name":"vol",
                    "ammount":{stats[11]}
                }},
                {{
                    "name":"car",
                    "ammount":{stats[12]}
                }},
                {{
                    "name":"sue",
                    "ammount":{stats[13]}
                }}
            ],
            "abilities":[]
        }},"""
        if(leyenda == -1):
            to_append += newmob

with open(actors_path,"a", encoding="utf-8", ) as mon_book:
    mon_book.write(to_append)