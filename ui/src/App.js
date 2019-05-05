import React, { Component } from 'react';
import moment from 'moment';
import { HashRouter as Router, Route, Link } from "react-router-dom";

var AppConfig = {
    API_BASE: window.location.protocol + '//' + window.location.host
};


var Consts = {
  IsAliveJob : 0,
  PortScanJob : 1,

  NormalScan : 0,
  SynScan : 1,
  FinScan : 2,

  Queued: 0,
  InProgress: 1,
  Completed: 2,

  PortOpen: 1 << 0,
  PortClosed: 1 << 1,
  PortFiltered: 1 << 2,
  PortUnFiltered: 1 << 3,

  IpAlive: 0,
  IpDead: 1,
};

function IpStatusStr(ipStatus) {
  ipStatus = parseInt(ipStatus);
  if (ipStatus === 0)
  {
    return "Alive"
  }

  if (ipStatus === 1)
  {
    return "Dead"
  }
  return "Unknown"
}

function TaskStateStr(taskState) {
  taskState = parseInt(taskState);
  if (taskState === 0)
  {
    return "Queued"
  }

  if (taskState === 1)
  {
    return "InProgress"
  }

  if (taskState === 2)
  {
    return "Completed"
  }

  return "Unknown"
}

function JobTypeStr(jobType) {
  jobType = parseInt(jobType);
  
  if (jobType === 0)
  {
    return "IsAlive"
  }

  if (jobType === 1)
  {
    return "PortScan"
  }

  return "Unknown"
}

function PortScanTypeStr(portScanType) {
  portScanType = parseInt(portScanType);
  if (portScanType === 0)
  {
    return "Normal"
  }

  if (portScanType === 1)
  {
    return "SYN"
  }

  if (portScanType === 2)
  {
    return "FIN"
  }

  return "Unknown"
}

function PortStatusStr(portStatus) {
  portStatus = parseInt(portStatus)
  let status = []
  if (portStatus & Consts.PortOpen)
  {
    status.push("PortOpen")
  }

  if (portStatus & Consts.PortClosed)
  {
    status.push("PortClosed")
  }

  if (portStatus & Consts.PortFiltered)
  {
    status.push("PortFiltered")
  }

  if (portStatus & Consts.PortUnFiltered)
  {
    status.push("PortUnFiltered")
  }

  return status.join(' | ')
}


class JobSnippet extends Component {
  deleteJob = (e) =>
    {
      let jobId = this.props.data.JobId;
      console.log(`delete job ${jobId}`)

      fetch(`${AppConfig.API_BASE}/jobs/${jobId}`, {
          method: 'delete'
        })
      .then(response => {
          if (response.ok) {
             this.props.removeRow(jobId);
             console.log('removed job');
          }
          else {
              console.log('Failed to delete job');
          }
      });

      e.preventDefault();
  }

  renderMain() {
    return  (<div className="job-snippet row mb-10 p-3">
      <div className="col-10">
      <div>Job <strong>#</strong>{this.props.data.JobId} /&nbsp;
      <strong>Type: </strong>{JobTypeStr(this.props.data.Type)} /&nbsp;
      <strong>Status: </strong>{TaskStateStr(this.props.data.JobState)} /&nbsp;
      <strong>Creation: </strong>{moment(this.props.data.JobCreatedAt).fromNow()} 
      {this.props.data.JobState == Consts.Completed ? (<span> / <strong>Duration:</strong> {moment.duration(moment(this.props.data.JobCompletedAt).diff(moment(this.props.data.JobCreatedAt))).asSeconds()} seconds</span>) : ''}
      </div>
      {this.props.data.Type == Consts.PortScanJob ? 
        (<div>Mode: {PortScanTypeStr(this.props.data.Data.Type)} / IP: {this.props.data.Data.Ip} / Ports: {this.props.data.Data.PortRange.Start}-{this.props.data.Data.PortRange.End}</div>)
        : (<div>IP / CIDR: {this.props.data.Data.IpBlock}</div>)}
        </div>
      {this.props.showButtons ? (<div className="col-2">
        <Link className="btn btn-primary mr-1" to={`/view/${this.props.data.JobId}`}>detail</Link>
        <button type="button" className="btn btn-danger mr-1" onClick={this.deleteJob}>delete</button>
      </div>) : <div/ >}
      </div>);
  }
  
  render() {
      return this.props.data.Data !== undefined ? this.renderMain() : (<div />);
  }
}

class PortScanResultView extends Component {
  render() {
    return (
        <table className="table table-striped">
        <thead>
          <tr className="">
            <th>Port</th>
            <th>Status</th>
            <th>Banner</th>
          </tr>
          </thead>
          <tbody>
          {
            Object.keys(this.props.data).map((key, index) => ( 
              <tr>
                <td>{key}</td>
                <td>{PortStatusStr(this.props.data[key].Status)}</td>
                <td>{`${this.props.data[key].hasOwnProperty('Banner') ? this.props.data[key].Banner : "Not available"}`}</td>
              </tr>
            ))
          }
            
          </tbody>
        </table>
      )
  }
}

class IsAliveResultView extends Component {
  render() {
    return (
        <table className="table table-striped">
        <thead>
          <tr className="">
            <th>Ip</th>
            <th>Status</th>
          </tr>
          </thead>
          <tbody>
          {
            this.props.data.map(result => 
              <tr>
                <td>{result.Ip}</td>
                <td>{IpStatusStr(result.Status)}</td>
              </tr>
            )
          }
            
          </tbody>
        </table>
      )
  }
}

class TaskView extends Component {
  constructor(props) {
    super(props);

    this.state = {
      visible: true
    }
  }
  
  toggleView = () => {
    this.setState({visible: !this.state.visible})
  }
  
  render() {
      return (
        <div className="card mb-2">
          <div className="card-header" style={{cursor: 'pointer'}} onClick={this.toggleView}>
             <span> Task <strong>#{this.props.data.TaskId}</strong> /&nbsp;
             <strong>State:</strong> {TaskStateStr(this.props.data.TaskState)} /&nbsp;
             <strong>Creation: </strong>{moment(this.props.data.CreateTime).format('LLL')} 
             {this.props.data.TaskState == Consts.Completed ? (<span> / <strong>Duration:</strong> {moment.duration(moment(this.props.data.CompleteTime).diff(moment(this.props.data.CreateTime))).asSeconds()} seconds</span>) : ''}
             </span>
             <div className={`${this.props.data.WorkerId == -1 ? 'd-none' : ''}`}>
               Worker <strong>#{this.props.data.WorkerId}</strong> /&nbsp;
               <strong>Name:</strong> {this.props.data.WorkerName} /&nbsp;
               <strong>Address:</strong> {this.props.data.WorkerAddress} 
             </div>
          </div>

          <div className={`card-body ${!this.state.visible ? 'd-none' : ''}`}>
           {this.props.data.TaskState == Consts.Completed ?
            (this.props.type == Consts.IsAliveJob ?
                (<IsAliveResultView data={this.props.data.Data} />) : <PortScanResultView data={this.props.data.Data} />) : ''}

          </div>
      </div>
        )
  }
}


class JobSubmit extends Component {
  constructor(props) {
    super(props);

    this.state = {
      jobType: 0,
      ipBlock: "",
      mode: 0,
      portMin: 0,
      portMax: 65535
    }
  }
  handleChange = (e) => {
        console.log(e.target.name, e.target.value);
        this.setState({[e.target.name] : e.target.value});
    };

  submitJob = (e) =>
  {
    console.log(`submit job`)
    let that = this;
    let params = {"Type": parseInt(this.state.jobType), "Data": {}};

    if (this.state.jobType == Consts.IsAliveJob)
    {
      params["Data"] = {
        "IpBlock": this.state.ipBlock
      };
    }
    else if (this.state.jobType == Consts.PortScanJob)
    {
      params["Data"] = {
        "Type": parseInt(this.state.mode),
        "Ip": this.state.ipBlock,
        "PortRange": {
          "Start": parseInt(this.state.portMin),
          "End": parseInt(this.state.portMax)
        }
      };
    }
    console.log(params)
    fetch(`${AppConfig.API_BASE}/jobs/`, {
        method: 'post',
        headers: {
        'Content-Type': 'application/json'
        },
        body: JSON.stringify(params)
      })
    .then(response => response.json().then(json => {
        if (response.ok) {
             params.JobId = json.JobId;
             params.JobState = Consts.Queued;
             params.JobCreatedAt = new Date
             that.props.appendRow(params);
        }
        else {
            console.log('Failed to submit job');
        }
    }));

    e.preventDefault();
}
  render() {
    return (
      <div className="p-3">
      <form>
          <div className="row form-inline mb-3">
           <div className="form-group col-2 mr-3">
           <label for="jobType" className="col-sm-2">Type</label>
            <div className="col-sm-4">
            <select style={{width:'120px'}} className="custom-select"
                    defaultValue="IsAlive"
                    name="jobType" onChange={this.handleChange}>
                <option value="0">IsAlive</option>
                <option value="1">PortScan</option>
            </select>
            </div>
            </div>
          <div className="form-group col-6 mr-2">
          <label for="ipBlock" className="col-sm-3" style={{marginRight:'-25px'}}>{`IP ${parseInt(this.state.jobType) == Consts.IsAliveJob ? 'or CIDR' : ''}`}</label>
          <div className="col-sm-4">
          <input style={{width:'215px'}}className="form-control"
                         type="text"
                         name="ipBlock"
                         placeholder="192.168.0.1"
                         onChange={this.handleChange} />
           </div>
           </div>
           </div>
          {this.state.jobType === "1" ?
          (<div className="row form-inline mb-3">
            <div className="form-group col-2 mr-3">
            <label for="mode" className="col-sm-2">Mode</label>
             <div className="col-sm-4">
            <select style={{width:'120px'}} className="custom-select"
                    defaultValue="NormalScan"
                    name="mode" onChange={this.handleChange}>
                <option value="0">Normal</option>
                <option value="1">SYN</option>
                <option value="2">FIN</option>
            </select>
            </div>
            </div>
            <div className="form-group col-6 mr-3">
            <label for="portMin" className="col-sm-3" style={{marginRight:'-25px'}}>Ports</label>
            <div className="col-sm-6">
            <input style={{width:'100px'}} className="form-control "
                           type="number"
                           name="portMin"
                           defaultValue="0"
                           onChange={this.handleChange} />
                           <span> - </span>
            <input style={{width:'100px'}} className="form-control"
                           type="number"
                           name="portMax"
                           defaultValue="443"
                           onChange={this.handleChange} /></div></div></div>) : '' }

          <button type="button" className="btn btn-primary" onClick={this.submitJob}>submit job</button>
      </form>
      </div>
    );
  }
}

class JobDetail extends Component {
  constructor(props) {
    super(props);
    this.state = {data: {}, job: {}};
  }


  componentDidMount() {

    fetch(`${AppConfig.API_BASE}/jobs/${this.props.match.params.id}`)
    .then(response => response.json().then( json => {
          if (response.ok) {
              console.log(json);
              this.setState({data: json});
              this.setState({ job: json.JobInfo});
          }
          else
          {
              console.log(JSON.stringify(json));
          }
      }));
  }
  
  render() {
    console.log(this.state);
    return (
      <div>
      <JobSnippet data={this.state.job} showButtons={false} />
      <div className="accordian">
          {this.state.data.Data && this.state.data.Data.map(task => (<TaskView type={this.state.data.JobInfo.Type} data={task} />))}
      </div>
      </div>
    );
  }
}

class JobList extends Component {
  constructor(props) {
    super(props);
    this.state = {data: []};
  }
  
  componentDidMount() {
    // fetch jobs here
    fetch(`${AppConfig.API_BASE}/jobs/`)
      .then(response => response.json().then( json => {
            if (response.ok) {
                console.log(json);
                this.setState({data: json.Jobs});
            }
            else
            {
                console.log(JSON.stringify(json));
            }
        }));
  }

  appendRow = (job) =>
  {
      let newitems = this.state.data;
      newitems.push(job);
      this.setState({
          data: newitems
      });
  };

  removeRow = (jobId) => {
      let newitems = this.state.data.filter(el => {
         return el.JobId != jobId;
      });
      
      this.setState({
          data: newitems
      });
  };
  
  render() {
    return (
      <div>
        <JobSubmit appendRow={this.appendRow} />
        <div>
            {this.state.data.map(job => (<JobSnippet data={job} removeRow={this.removeRow} showButtons={true} />))}
        </div>
      </div>
    );
  }
}

function Home() {
  return <JobList />;
}

function About() {
  return <div>
        <h3 className="mt-3">About</h3>
        <p>
          <strong>Team:</strong> Jarvis <br /><br />

          <ol>
            <li>Akshay Aurora</li>
            <li>Rory Bennett</li>
            <li>Jarin Moon</li>
          </ol>
        </p>
        </div>;
}

class App extends Component {
  render() {
    return (
      <div className="App">
      <Router>
        <nav className="navbar navbar-expand-lg navbar-light bg-light">
          <span className="navbar-brand mb-0 h1">Jarvis Scanner</span>
            <ul className="navbar-nav mr-auto">
             <li className="nav-item active">
                <Link className="nav-link active" to='/'>Home</Link>
             </li>
            <li className="nav-item">
                <Link className="nav-link" to='/about'>About</Link>
            </li>
                
            </ul>
        </nav>
        <div className="container">
          <Route path="/" exact component={Home} />
          <Route path="/about/" component={About} />
          <Route path='/view/:id' component={JobDetail} />
        </div>
      </Router>
      </div>
    );
  }
}

export default App;
