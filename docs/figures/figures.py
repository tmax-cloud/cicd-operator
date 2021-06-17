from urllib.request import urlretrieve

from diagrams import Diagram, Cluster, Edge
from diagrams.onprem.vcs import Github, Gitlab
from diagrams.k8s.others import CRD
from diagrams.k8s.compute import Pod
from diagrams.k8s.network import Ingress
from diagrams.custom import Custom

attr = {
    "pad": "0.5",
    "fontname": "Open Sans",
    "labelfontname": "Open Sans",
}

fontAttr = {
    "fontname": "Open Sans",
    "labelfontname": "Open Sans",
}

with Diagram("Architecture", show=False, graph_attr=attr, node_attr=fontAttr):
    tektonIconURL = "https://cd.foundation/wp-content/uploads/sites/78/2020/04/tekton-icon-color-1.png"
    tektonIcon = "tekton.png"
    urlretrieve(tektonIconURL, tektonIcon)

    with Cluster("Remote Git"):
        github = Github("github")
        gitlab = Gitlab("gitlab")
        gits = [
            github,
            gitlab]

    with Cluster("K8s cluster"):
        ic = CRD("IntegrationConfig")
        ij = CRD("IntegrationJob")

        ing = Ingress()

        with Cluster("CI/CD Operator"):
            webhookServer = Pod("Webhook server")

        github - Edge(label="PullRequest/Push/Tag") >> ing >> webhookServer
        webhookServer - Edge(label="Queued") >> ij

        pr = Custom("PipelineRun", tektonIcon)
        ij - Edge(label="Scheduled") >> pr

        tr1 = Custom("TaskRun-1", tektonIcon)
        tr2 = Custom("TaskRun-2", tektonIcon)
        tr3 = Custom("TaskRun-3", tektonIcon)

        pr >> tr3 >> Pod()
        pr >> tr2 >> Pod()
        pr >> tr1 >> Pod()

        ic - Edge(label="Register Webhook") >> gitlab
