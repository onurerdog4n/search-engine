import json
import random
from datetime import datetime, timedelta

def generate_json_data(count=100):
    contents = []
    tags_pool = ["programming", "tutorial", "go", "docker", "cloud", "backend", "concurrency", "performance", "testing"]
    
    start_date = datetime(2026, 1, 1)
    
    for i in range(1, count + 1):
        published_at = (start_date + timedelta(days=random.randint(0, 31))).isoformat() + "Z"
        content_type = "video" if i % 3 != 0 else "article"
        
        metrics = {}
        if content_type == "video":
            metrics = {
                "views": random.randint(1000, 50000),
                "likes": random.randint(100, 5000),
                "duration": f"{random.randint(5, 45)}:{random.randint(10, 59):02d}"
            }
        else:
            metrics = {
                "reading_time": random.randint(5, 20),
                "reactions": random.randint(50, 600)
            }

        content = {
            "id": f"json-{'v' if content_type == 'video' else 'a'}{i}",
            "title": f"JSON Content {i}: Learning {random.choice(tags_pool).capitalize()}",
            "type": content_type,
            "metrics": metrics,
            "published_at": published_at,
            "tags": random.sample(tags_pool, k=random.randint(1, 3))
        }
        contents.append(content)
        
    response = {
        "contents": contents
    }
    
    with open("mocks/provider1.json", "w") as f:
        json.dump(response, f, indent=2)

def generate_xml_data(count=100):
    categories_pool = ["devops", "kubernetes", "ci-cd", "cloud", "security", "monitoring", "architecture", "programming"]
    start_date = datetime(2026, 1, 1)
    
    xml_output = ['<?xml version="1.0" encoding="UTF-8"?>', '<feed>', '  <items>']
    
    for i in range(1, count + 1):
        pub_date = (start_date + timedelta(days=random.randint(0, 31))).strftime("%Y-%m-%d")
        content_type = "video" if i % 4 != 0 else "article"
        
        xml_output.append('    <item>')
        xml_output.append(f'      <id>xml-{"v" if content_type == "video" else "a"}{i}</id>')
        xml_output.append(f'      <headline>XML Content {i}: Modern {random.choice(categories_pool).capitalize()} Guide</headline>')
        xml_output.append(f'      <type>{content_type}</type>')
        xml_output.append('      <stats>')
        
        if content_type == "video":
            xml_output.append(f'        <views>{random.randint(5000, 30000)}</views>')
            xml_output.append(f'        <likes>{random.randint(200, 2000)}</likes>')
            xml_output.append(f'        <duration>{random.randint(10, 60)}:{random.randint(10, 59):02d}</duration>')
        else:
            xml_output.append(f'        <reading_time>{random.randint(5, 20)}</reading_time>')
            xml_output.append(f'        <reactions>{random.randint(50, 600)}</reactions>')
            xml_output.append(f'        <comments>{random.randint(5, 100)}</comments>')
            
        xml_output.append('      </stats>')
        xml_output.append(f'      <publication_date>{pub_date}</publication_date>')
        xml_output.append('      <categories>')
        for cat in random.sample(categories_pool, k=random.randint(1, 2)):
            xml_output.append(f'        <category>{cat}</category>')
        xml_output.append('      </categories>')
        xml_output.append('    </item>')
        
    xml_output.append('  </items>')
    xml_output.append('</feed>')
    
    with open("mocks/provider2.xml", "w") as f:
        f.write("\n".join(xml_output))

if __name__ == "__main__":
    generate_json_data(120)
    generate_xml_data(115)
    print("Generated provider1.json (120 items) and provider2.xml (115 items)")
