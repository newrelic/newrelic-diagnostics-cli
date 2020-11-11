"""
This script takes csv formatted task dependency output from the Diagnostics CLI
and creates a dependency graph using the graphviz library. 
It expects a csv list from STDIN in the form: dependency, parent_task.
This script is meant to be fed the output of `newrelic-diagnostics  --print-dependencies`
Tasks are expected to be in the format "Category/Subcategory/Taskname"
"""

from graphviz import Digraph
import sys
import csv

color_map = {
    'python': 'darkgoldenrod1',
    'java': 'burlywood',
    'ruby': 'indianred',
    'dotnetcore': 'skyblue',
    'dotnet':'skyblue3',
    'base':'gray',
    'php':'mediumslateblue',
    'synthetics':'purple1',
    'infra':'cyan',
    'node':'chartreuse3',
    'ios':'hotpink',
    'android':'darkolivegreen1',
    'browser': 'red',

}

node_opts = {
    'shape' : 'record',
    'style' : 'filled',
    'color' : 'black',
    'fontsize':'12',
    'fontname' : 'arial',
    'penwidth' : '3',
}

graph_opts = {
    'dpi' : '400',
    'ranksep': '1',
}

def beautify_task(cat, subcat, name): 
    """HTML format NRDiag task elements for pretty graphing"""
    label_template = '<<FONT FACE="arial bold">{cat}</FONT><br/>{subcat}<br/>{name}>'
    return label_template.format(cat=cat, subcat=subcat, name=name)

f = Digraph('dependencies', format='png', engine='dot', graph_attr=graph_opts)
f.attr(size='10,10')
f.node_attr.update(color='lightblue2', style='filled')

csv_reader = csv.reader(sys.stdin.readlines(), delimiter=',')

for row in csv_reader:

    for node in row:
        if node:
            cat, subcat, name = node.split('/')
            node_label = beautify_task(cat, subcat, name)
            node_fill = color_map.get(cat.lower(), 'white')
            f.node(node, fillcolor=node_fill, label=node_label, **node_opts)
    
    if row[0]:
        # draw an edge
        target_category = row[1].split('/')[0]
        edge_color = color_map.get(target_category.lower(), 'black')
        f.edge(*row, color=edge_color, penwidth=node_opts['penwidth'])

f.render('docs/images/dependencies', cleanup=True)