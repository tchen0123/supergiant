package core

import (
	"fmt"
	"strings"
	"time"

	"github.com/supergiant/guber"
	"github.com/supergiant/supergiant/common"
)

type instanceType struct {
	ID    string
	Name  string
	RAM   float64
	Cores float64
	Cost  float64 // Linux On Demand Hourly
}

// These are sorted by cost ascending
var instanceTypes = [...]*instanceType{
	&instanceType{"t2.nano", "T2 Nano", 0.5, 1.0, 0.006},
	&instanceType{"t2.micro", "T2 Micro", 1.0, 1.0, 0.013},
	&instanceType{"t1.micro", "T1 Micro", 0.613, 1.0, 0.020},
	&instanceType{"t2.small", "T2 Small", 2.0, 1.0, 0.026},
	&instanceType{"m1.small", "M1 General Purpose Small", 1.7, 1.0, 0.044},
	&instanceType{"t2.medium", "T2 Medium", 4.0, 2.0, 0.052},
	&instanceType{"m3.medium", "M3 General Purpose Medium", 3.75, 1.0, 0.067},
	&instanceType{"m1.medium", "M1 General Purpose Medium", 3.75, 1.0, 0.087},
	&instanceType{"t2.large", "T2 Large", 8.0, 2.0, 0.104},
	&instanceType{"c3.large", "C3 High-CPU Large", 3.75, 2.0, 0.105},
	&instanceType{"c4.large", "C4 High-CPU Large", 3.75, 2.0, 0.105},
	&instanceType{"m4.large", "M4 Large", 8.0, 2.0, 0.120},
	&instanceType{"c1.medium", "C1 High-CPU Medium", 1.7, 2.0, 0.130},
	&instanceType{"m3.large", "M3 General Purpose Large", 7.5, 2.0, 0.133},
	&instanceType{"r3.large", "R3 High-Memory Large", 15.25, 2.0, 0.166},
	&instanceType{"m1.large", "M1 General Purpose Large", 7.5, 2.0, 0.175},
	&instanceType{"c4.xlarge", "C4 High-CPU Extra Large", 7.5, 4.0, 0.209},
	&instanceType{"c3.xlarge", "C3 High-CPU Extra Large", 7.5, 4.0, 0.210},
	&instanceType{"m4.xlarge", "M4 Extra Large", 16.0, 4.0, 0.239},
	&instanceType{"m2.xlarge", "M2 High Memory Extra Large", 17.1, 2.0, 0.245},
	&instanceType{"m3.xlarge", "M3 General Purpose Extra Large", 15.0, 4.0, 0.266},
	&instanceType{"r3.xlarge", "R3 High-Memory Extra Large", 30.5, 4.0, 0.333},
	&instanceType{"m1.xlarge", "M1 General Purpose Extra Large", 15.0, 4.0, 0.350},
	&instanceType{"c4.2xlarge", "C4 High-CPU Double Extra Large", 15.0, 8.0, 0.419},
	&instanceType{"c3.2xlarge", "C3 High-CPU Double Extra Large", 15.0, 8.0, 0.420},
	&instanceType{"m4.2xlarge", "M4 Double Extra Large", 32.0, 8.0, 0.479},
	&instanceType{"m2.2xlarge", "M2 High Memory Double Extra Large", 34.2, 4.0, 0.490},
	&instanceType{"c1.xlarge", "C1 High-CPU Extra Large", 7.0, 8.0, 0.520},
	&instanceType{"m3.2xlarge", "M3 General Purpose Double Extra Large", 30.0, 8.0, 0.532},
	&instanceType{"g2.2xlarge", "G2 Double Extra Large", 15.0, 8.0, 0.650},
	&instanceType{"r3.2xlarge", "R3 High-Memory Double Extra Large", 61.0, 8.0, 0.665},
	&instanceType{"d2.xlarge", "D2 Extra Large", 30.5, 4.0, 0.690},
	&instanceType{"c4.4xlarge", "C4 High-CPU Quadruple Extra Large", 30.0, 16.0, 0.838},
	&instanceType{"c3.4xlarge", "C3 High-CPU Quadruple Extra Large", 30.0, 16.0, 0.840},
	&instanceType{"i2.xlarge", "I2 Extra Large", 30.5, 4.0, 0.853},
	&instanceType{"m4.4xlarge", "M4 Quadruple Extra Large", 64.0, 16.0, 0.958},
	&instanceType{"m2.4xlarge", "M2 High Memory Quadruple Extra Large", 68.4, 8.0, 0.980},
	&instanceType{"r3.4xlarge", "R3 High-Memory Quadruple Extra Large", 122.0, 16.0, 1.330},
	&instanceType{"d2.2xlarge", "D2 Double Extra Large", 61.0, 8.0, 1.380},
	&instanceType{"c4.8xlarge", "C4 High-CPU Eight Extra Large", 60.0, 36.0, 1.675},
	&instanceType{"c3.8xlarge", "C3 High-CPU Eight Extra Large", 60.0, 32.0, 1.680},
	&instanceType{"i2.2xlarge", "I2 Double Extra Large", 61.0, 8.0, 1.705},
	&instanceType{"cc2.8xlarge", "Cluster Compute Eight Extra Large", 60.5, 32.0, 2.000},
	&instanceType{"cg1.4xlarge", "Cluster GPU Quadruple Extra Large", 22.5, 16.0, 2.100},
	&instanceType{"m4.10xlarge", "M4 Deca Extra Large", 160.0, 40.0, 2.394},
	&instanceType{"g2.8xlarge", "G2 Eight Extra Large", 60.0, 32.0, 2.600},
	&instanceType{"r3.8xlarge", "R3 High-Memory Eight Extra Large", 244.0, 32.0, 2.660},
	&instanceType{"d2.4xlarge", "D2 Quadruple Extra Large", 122.0, 16.0, 2.760},
	&instanceType{"hi1.4xlarge", "HI1. High I/O Quadruple Extra Large", 60.5, 16.0, 3.100},
	&instanceType{"i2.4xlarge", "I2 Quadruple Extra Large", 122.0, 16.0, 3.410},
	&instanceType{"cr1.8xlarge", "High Memory Cluster Eight Extra Large", 244.0, 32.0, 3.500},
	&instanceType{"hs1.8xlarge", "High Storage Eight Extra Large", 117.0, 16.0, 4.600},
	&instanceType{"d2.8xlarge", "D2 Eight Extra Large", 244.0, 36.0, 5.520},
	&instanceType{"i2.8xlarge", "I2 Eight Extra Large", 244.0, 32.0, 6.820},
}

var (
	waitBeforeScale         = 2 * time.Minute
	minAgeToExist           = 10 * time.Minute // this is used to prevent adding more nodes while still-pending pods are scheduling to a new node
	maxClusteredPodsPerNode = 2                // prevent putting all nodes of a cluster on one host node
	maxDisksPerNode         = 11
	trackedEventMessages    = [...]string{
		"MatchNodeSelector",
		"PodExceedsMaxPodNumber",
		"PodExceedsFreeMemory",
		"PodExceedsFreeCPU",
		"no nodes available to schedule pods",
	}
)

type capacityService struct {
	core                *Core
	instanceTypes       []*instanceType
	largestInstanceType *instanceType
}

func newCapacityService(c *Core) *capacityService {
	s := new(capacityService)
	s.core = c

	instanceTypeIDs, err := autoscalingGroupInstanceTypes(c)
	if err != nil {
		panic(err)
	}

	for _, instanceTypeID := range instanceTypeIDs {
		for _, it := range instanceTypes {
			if it.ID == instanceTypeID {

				Log.Infof("Capacity service registered AWS instance type %s", it.ID)

				s.instanceTypes = append(s.instanceTypes, it)
				break
			}
		}
	}

	s.largestInstanceType = s.instanceTypes[len(s.instanceTypes)-1]

	return s
}

type projectedNode struct {
	Committed bool
	Size      *instanceType
	Pods      []*guber.Pod
}

func (pnode *projectedNode) usedRAM() (u float64) {
	for _, pod := range pnode.Pods {
		for _, container := range pod.Spec.Containers {

			// NOTE we use limits here, and not requests, because we want to spin up
			// nodes that are at least slightly bigger than the user thinks the pod
			// could utilize. This will ensure that the user's limit CAN BE FILLED AT
			// ALL. This is at the core of our increased-utilization strategy.

			memStr := container.Resources.Limits.Memory
			b := new(common.BytesValue)
			if err := b.UnmarshalJSON([]byte(memStr)); err != nil {
				panic(err)
			}
			u += b.Gibibytes()
		}
	}
	return
}

func (pnode *projectedNode) usedCPU() (u float64) {
	for _, pod := range pnode.Pods {
		for _, container := range pod.Spec.Containers {

			// NOTE above in usedRAM

			cpuStr := container.Resources.Limits.CPU
			c := new(common.CoresValue)
			if err := c.UnmarshalJSON([]byte(cpuStr)); err != nil {
				panic(err)
			}
			u += c.Cores()
		}
	}
	return
}

func (pnode *projectedNode) usedVolumes() (u int) {
	for _, pod := range pnode.Pods {
		u += len(pod.Spec.Volumes)
	}
	return
}

func (pnode1 *projectedNode) canMergeWith(pnode2 *projectedNode) bool {
	usedCPU := pnode1.usedCPU() + pnode2.usedCPU()
	usedRAM := pnode1.usedRAM() + pnode2.usedRAM()
	usedVolumes := pnode1.usedVolumes() + pnode2.usedVolumes()
	return pnode1.Size.Cores >= usedCPU && pnode1.Size.RAM >= usedRAM && usedVolumes <= maxDisksPerNode
}

//------------------------------------------------------------------------------

func (s *capacityService) hasTrackedEvent(pod *guber.Pod) (bool, error) {
	q := &guber.QueryParams{
		FieldSelector: "involvedObject.name=" + pod.Metadata.Name,
	}
	events, err := s.core.k8s.Events("").Query(q)
	if err != nil {
		return false, err
	}

	for _, event := range events.Items {
		for _, message := range trackedEventMessages {
			if strings.Contains(event.Message, message) {
				return true, nil
			}
		}
	}
	return false, nil
}

func (s *capacityService) incomingPods() (incomingPods []*guber.Pod, err error) {
	waitStart := time.Now()

	for {
		incomingPods = incomingPods[:0] // reset

		q := &guber.QueryParams{
			FieldSelector: "status.phase=Pending",
		}
		// TODO does this get all pods?
		pendingPods, err := s.core.k8s.Pods("").Query(q)
		if err != nil {
			return nil, err
		}

		for _, pod := range pendingPods.Items {
			hasTrackedEvent, err := s.hasTrackedEvent(pod)
			if err != nil {
				return nil, err
			}
			if hasTrackedEvent {
				incomingPods = append(incomingPods, pod)
			}
		}

		elapsed := time.Since(waitStart)
		incomingCount := len(incomingPods)

		if incomingCount > 0 && elapsed < waitBeforeScale {

			Log.Debugf("Waiting to add nodes for %d pods", incomingCount)

			time.Sleep(2 * time.Second)
		} else {
			break
		}
	}

	return incomingPods, nil
}

func (s *capacityService) Run() {
	for _ = range time.NewTicker(1 * time.Second).C {

		// Log.Debug("Capacity service loop")

		incomingPods, err := s.incomingPods()
		if err != nil {
			Log.Errorf("Capacity service error when fetching incoming pods: %s", err)
			continue
		}

		var projectedNodes []*projectedNode
		for _, pod := range incomingPods {
			projectedNodes = append(projectedNodes, &projectedNode{
				false,
				s.largestInstanceType,
				[]*guber.Pod{pod},
			})
		}

		for {
			var (
				pnode1      *projectedNode
				pnode2      *projectedNode
				pnode2Index int
			)

			for _, pnode := range projectedNodes {
				fmt.Println("-------------------------------------------------------------")
				fmt.Println(pnode)
				for _, pod := range pnode.Pods {
					fmt.Println(pod)
				}
				fmt.Println("-------------------------------------------------------------")
			}

			//==========================================================================
			// find an uncommitted nodeAndPod
			//==========================================================================

			for _, pnode := range projectedNodes {
				if !pnode.Committed {
					pnode1 = pnode
					break
				}
			}

			if pnode1 == nil {
				break
			}

			//==========================================================================
			// find a pnode2 you can merge pnode1 with
			//==========================================================================

			for pnode2IndexCandidate, pnode2Candidate := range projectedNodes {
				if pnode2Candidate == pnode1 { // don't want to merge with self
					continue
				}

				if pnode1.canMergeWith(pnode2Candidate) {
					pnode2 = pnode2Candidate
					pnode2Index = pnode2IndexCandidate
					break
				}
			}

			//==========================================================================
			// merge if found, OR scale down to the smallest instance size it can use and commit it
			//==========================================================================

			if pnode2 != nil {
				// Delete the partner being merged, and merge pods
				i := pnode2Index
				projectedNodes = append(projectedNodes[:i], projectedNodes[i+1:]...)
				pnode1.Pods = append(pnode1.Pods, pnode2.Pods...)
			} else {
				// If we can't merge with anyone, can we scale down to the lowest cost.
				// instanceTypes are asc. by cost, so the first we find is the cheapest.
				for _, instanceType := range s.instanceTypes {
					if instanceType.Cores >= pnode1.usedCPU() && instanceType.RAM >= pnode1.usedRAM() {
						pnode1.Size = instanceType
						pnode1.Committed = true
						break
					}

				}
			}

		}

		// if err := s.core.Nodes().populate(); err != nil {
		// 	Log.Errorf("Capacity service error when populating Nodes: %s", err)
		// }

		existingNodes, err := s.core.Nodes().List()
		if err != nil {
			Log.Errorf("Capacity service error when fetching existing Nodes: %s", err)
		}

		for _, node := range existingNodes.Items {

			// we can't do "has pods with volumes", tho..... but can we do something like..... has pods with any resource request?
			// TODO ---- need to label them to prevent disk overflow

			// eventual option to delete nodes when there are pods (w/ or wo/ volumes?) that could move to other nodes (we would have to calculate that)

			hasPods, err := node.hasPodsWithReservedResources()
			if err != nil {
				Log.Errorf("Capacity service error when fetching Pods for Node: %s", err)
			}

			if !hasPods && time.Since(node.ProviderCreationTimestamp.Time) > minAgeToExist {

				Log.Infof("Terminating node %s", node.Name)

				if err := node.Delete(); err != nil {
					Log.Errorf("Capacity service error when deleting Node: %s", err)
				}
			}
		}

		for _, pnode := range projectedNodes {
			node := &NodeResource{
				Node: &common.Node{
					Class: pnode.Size.ID,
				},
			}

			// If there's an existing node which is spinning up with this type, then
			// don't create.
			// This is a big TODO -- the logic should be much tighter, allowing on
			// going projection of pods onto nodes that are still spinning up.

			alreadySpinningUp := false
			for _, existingNode := range existingNodes.Items {
				if existingNode.Class == node.Class && (existingNode.Status == "NOT_READY" || time.Since(node.ProviderCreationTimestamp.Time) < minAgeToExist) {
					// This may be a node that is already being created, or NOTE it could
					// be a broken node that we erroneously identify as spinning up.
					alreadySpinningUp = true
					break
				}
			}
			if alreadySpinningUp {
				continue
			}

			if err := s.core.Nodes().Create(node); err != nil {
				Log.Errorf("Capacity service error when creating Node: %s", err)
			}
		}
	}
}
